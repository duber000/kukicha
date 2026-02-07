# FAQ (Frequently Asked Questions)

1. Why use Kukicha instead of just using Python?

Great question! While Kukicha was designed to feel familiar to Python developers, the underlying "soul" is entirely different. Python is an interpreted, dynamic language, whereas Kukicha is a statically-typed, compiled language that transpiles directly to Go.

### What Feels Familiar

If you know Python, you already know a lot of Kukicha:

| Python | Kukicha | Notes |
|--------|---------|-------|
| `and`, `or`, `not` | `and`, `or`, `not` | Identical! |
| `if x == y:` | `if x equals y` | English keyword |
| `for x in items:` | `for x in items` | Same iteration style |
| `# comment` | `# comment` | Identical! |
| Indentation (4 spaces) | Indentation (4 spaces) | Identical! |
| f-strings `f"{name}"` | `"{name}"` | No prefix needed |
| `def greet(name):` | `func Greet(name string)` | Types required |
| Default params | `func F(x int = 10)` | Same concept |
| `**kwargs` / named args | `F(x: 10)` | Clean syntax |

### Key Differences to Learn

Python developers should be aware of these key distinctions:

1. **Static Types**: Function parameters require explicit types. Local variables are inferred.
   ```kukicha
   func Greet(name string)      # 'name' must specify type
       message := "Hi"          # 'message' type is inferred
   ```

2. **No Implicit Returns**: Use `return` explicitly.

3. **Error Handling**: Python uses exceptions; Kukicha uses `onerr` for explicit handling.
   ```kukicha
   data := files.Read("config.json") onerr panic "failed"
   ```

4. **The Pipe Operator**: Chain operations in a data-flow style (think `|` in shell).
   ```kukicha
   result := text |> string.TrimSpace() |> string.ToLower()
   ```

### Why Make the Switch?

Python is improving rapidly — Python 3.13+ has experimental free-threading (no GIL), and tools like **uv** make dependency management much smoother. So why Kukicha?

| Consideration | Python (2024+) | Kukicha |
|---------------|----------------|---------|
| **Deployment** | Much better with `uv` + containers | Single static binary, truly zero dependencies |
| **Concurrency** | Free-threading is experimental; async still has limits | Goroutines are production-proven, trivial to use |
| **Type Safety** | Optional (mypy, pyright) — runtime errors still possible | Mandatory compile-time checking |
| **Performance** | Slow for CPU-bound tasks even without GIL | 10-100x faster, compiles to native code |
| **Ecosystem** | Massive, unparalleled for ML/data science | Full access to Go ecosystem + any Go library |

**Bottom line**: If you're happy with Python and your workloads are I/O-bound or ML-focused, stay with Python! Kukicha shines for **systems programming, DevOps, CLI tools, and high-performance services** where you want Go's speed and deployment story with a friendlier syntax.

### Transition Tips

1. **Start with the [Beginner Tutorial](tutorial/beginner-tutorial.md)** — it's written for Python developers.
2. **Use `kukicha run`** to iterate quickly, just like running `python script.py`.
3. **Embrace the pipe operator** — it replaces many for-loops and makes data transformations readable.
4. **Don't fight the types** — they catch bugs before runtime.

2. Doesn't XGo (formerly Go+) already do this?

XGo is an excellent project, but it serves a different niche. 

    Semantic Keywords: Kukicha replaces symbols with English words (e.g., reference of user instead of &user, or and/or instead of &&/||). Go+ stays closer to standard Go syntax.

    The Pipe Operator: Kukicha is built around a "Data-Flow" philosophy. Our Smart Pipe (|>) logic allows you to chain complex operations with placeholders (_), making it significantly more readable for DevOps and API logic.

    Error Handling: Kukicha's onerr keyword is a unique middle ground—it removes the "if err != nil" boilerplate without hiding errors behind magic exceptions.

3. How do pipes handle Go's "Writer-First" or "Context-First" APIs?

Standard Go often places a `context.Context` or an `io.Writer` as the first argument. Kukicha handles this with **Smart Pipe Logic**:

1. **Automatic Context Handling**: If you pipe a `context.Context` (or any variable named `ctx`), it is automatically prepended as the first argument:
   ```kukicha
   ctx |> db.FetchUser(userID)  # Becomes: db.FetchUser(ctx, userID)
   ```

2. **Explicit Placeholders**: Use `_` to specify exactly where data should go (crucial for Writers):
   ```kukicha
   todo |> json.MarshalWrite(response, _)  # Becomes: json.MarshalWrite(response, todo)
   ```

4. Does Kukicha have a runtime?

No. Kukicha has zero runtime overhead. The compiler transpiles your code into standard, idiomatic Go. Once compiled by the Go toolchain, there is no trace of Kukicha left—just a high-performance Go binary.

5. Can I use existing Go libraries?

Yes. You can import any Go package (standard library or third-party) and use it directly in Kukicha. If the compiler hasn't seen the type before, it "trusts" the external package, allowing you to leverage the entire Go ecosystem immediately.

6. Does Kukicha support named arguments or default parameters?

Yes! Kukicha has both features to make code more readable:

**Default Parameters** let you specify default values for function parameters:
```kukicha
func Greet(name string, greeting string = "Hello")
    print("{greeting}, {name}!")

Greet("Alice")          # "Hello, Alice!"
Greet("Bob", "Hi")      # "Hi, Bob!"
```

**Named Arguments** let you specify argument names at the call site:
```kukicha
func Connect(host string, port int = 8080, timeout int = 30)
    # ...

# Clear and self-documenting
Connect("localhost", timeout: 60)
Connect("api.example.com", port: 443, timeout: 120)
```

Note: Named arguments must come after positional arguments, and parameters with defaults must come after those without.

7. What editor support is available?

**Zed** is currently supported with full language support:

- **Syntax Highlighting**: Tree-sitter grammar for accurate highlighting of Kukicha syntax
- **Diagnostics**: Real-time error reporting from the parser and semantic analyzer
- **Hover Information**: Type info for functions, types, interfaces, and builtins
- **Go-to-Definition**: Jump to function, type, interface, and field definitions
- **Code Completions**: Keywords, builtins, types, and declarations
- **Document Symbols**: Outline view of your code structure

**Installation:**

```bash
# 1. Build and install the LSP server
make install-lsp

# 2. Ensure GOPATH/bin is in your PATH
export PATH="$PATH:$(go env GOPATH)/bin"

# 3. Install the Zed extension (from the kukicha repo)
# In Zed, run: "zed: install dev extension"
# Select the editors/zed directory
```

The extension includes both the Tree-sitter grammar for syntax highlighting and integration with the `kukicha-lsp` language server for IDE features.
