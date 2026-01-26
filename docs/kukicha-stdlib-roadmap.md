# Kukicha Standard Library Roadmap

**Version:** 3.0.0
**Status:** Scripting-Focused
**Updated:** 2026-01-23

---

## Design Philosophy: "Go for Scripts"

Kukicha combines two powerful ideas:

1. **"It's Just Go"** - Use any Go package directly with `onerr` for error handling
2. **"Scripting Superpowers"** - Add high-level helpers for common scripting tasks

**Key Principles:**
- Go stdlib is first-class in Kukicha (no wrappers!)
- Pipe operator `|>` makes data transformations readable
- Kukicha stdlib provides scripting conveniences Go lacks
- Perfect for one-off tools, automation scripts, and learning
- All examples showcase pipe-based workflows

**Target Use Cases:**
- Fetching and processing API data
- File manipulation and text processing
- Building CLI tools quickly
- Automation scripts (better than bash)
- Learning programming concepts

---

## Implementation Status

### Quick Summary

**Ready to Use:** ✅ iter, slice, string, files, json, parse (CSV/YAML), fetch, concurrent, shell, cli, http (basic), template, result

**Partially Implemented:** ⚠️ retry (limited stub due to error handling design constraints)

**Limitations:**
- Constants with types not supported
- Retry package has limited functionality - see package documentation

### ✅ Completed Packages

| Package | Purpose | Status | Functions |
|---------|---------|--------|-----------|
| **iter** | Functional iteration (Filter, Map, Reduce) | ✅ Ready | Filter, Map, FlatMap, Take, Skip, Reduce, Collect, Find, Any, All, Enumerate, Zip, Chunk |
| **slice** | Slice operations with generics | ✅ Ready | First, Last, Drop, DropLast, Reverse, Unique, Chunk, Filter, Map, Contains, IndexOf, Concat, GroupBy |
| **string** | String utilities | ✅ Ready | ToUpper, ToLower, Title, Trim, TrimSpace, TrimPrefix, TrimSuffix, Split, Join, Contains, HasPrefix, HasSuffix, Index, Count, Replace, ReplaceAll, and more |
| **files** | File operations with pipes | ✅ Ready | Read, Write, Append, Exists, IsDir, IsFile, List, Delete, Copy, Move, MkDir, TempFile, TempDir, Size, ModTime, Extension, Join, Abs, Watch, UseWith |
| **json** | Pipe-friendly jsonv2 wrapper | ✅ Ready | NewEncoder, WithDeterministic, WithIndent, Encode, NewDecoder, Decode, Marshal, MarshalPretty, Unmarshal, MarshalWrite, UnmarshalRead |
| **parse** | CSV/YAML parsing | ✅ Ready | JsonPretty, Csv, CsvWithHeader, YamlPretty |
| **fetch** | HTTP client with json integration | ✅ Ready | Get, Post, New, Header, Timeout, Method, Body, Do, CheckStatus, Text, Bytes |
| **concurrent** | Concurrency helpers | ✅ Ready | Parallel, ParallelWithLimit, Go |
| **shell** | Safe command execution | ✅ Ready | Builder: New, Dir, SetTimeout, Env, Execute, Success, GetOutput, GetError, ExitCode. Direct: Run, RunSimple, RunWithDir, Which, Getenv, Setenv, Environ |
| **cli** | CLI argument parsing | ✅ Ready | Builder: New, Arg, AddFlag, Action, RunApp, GetString, GetBool, GetInt. Direct: Parse, String, Command, Flag, BoolFlag, IntFlag, PrintUsage |
| **http** | HTTP server helpers | ✅ Ready | WithCSRF, Serve |
| **template** | Text templating | ✅ Ready | Render, Data, Execute, Parse, New, WithContent, RenderSimple, Must |
| **result** | Optional/Result types | ✅ Ready | Some, None, Ok, Err, Map, UnwrapOr, AndThen, Match, ToOptional, FromOptional, Flatten, FlattenResult, All, Any |

### 🚧 Future Enhancements

| Package | Purpose | Status | Notes |
|---------|---------|--------|-------|
| **retry** | Full retry logic with automatic error handling | ⚠️ Limited | Currently a stub with helper functions. Full implementation requires language support for passing functions that return errors. See stdlib/retry/retry.kuki for manual retry patterns. |

### ⚠️ Known Limitations in Examples

Some roadmap examples use aspirational syntax not yet supported:

| Feature | What's Not Supported | Workaround |
|---------|---------------------|------------|
| **retry** | Automatic retry with `retry.Do()` | Use manual retry loops (see stdlib/retry/retry.kuki) |
| **shell** | Command piping with `shell.Pipe()` | Use shell pipes directly or multiple commands |

---

### Iter Package

Functional iteration with lazy evaluation and pipes:

```kukicha
import "stdlib/iter"
import "stdlib/slice"

# Pipeline: filter positive numbers, double them, sum
total := numbers
    |> slice.Filter(func(n int) bool {
        return n > 0
    })
    |> slice.Map(func(n int) int {
        return n * 2
    })
    |> iter.Reduce(0, func(acc int, n int) int {
        return acc + n
    })

# Find first matching item
user := users
    |> slice.Filter(func(u User) bool {
        return u.Email equals "admin@example.com"
    })
    |> slice.First(1)

# Take first 10, skip first 2
page := items
    |> slice.Drop(2)
    |> slice.First(10)

# Available functions:
# Filter, Map, FlatMap, Take, Skip, Enumerate, Zip
# Chunk, Reduce, Collect, Any, All, Find
```

### Slice Package

Pipeline-friendly slice operations:

```kukicha
import "stdlib/slice"

# Clean and process data pipeline
cleaned := rawData
    |> slice.Drop(1)           # Remove header
    |> slice.DropLast(1)       # Remove footer
    |> slice.Filter(isValid)   # Keep valid items
    |> slice.Unique()          # Remove duplicates
    |> slice.Reverse()         # Newest first

# Extract and transform
ids := users
    |> slice.Map(func(u User) int {
        return u.ID
    })
    |> slice.First(10)

# Batch processing
batches := items
    |> slice.Chunk(100)
    |> slice.Map(processBatch)

# Group items by category (Go 1.25+ generics with comparable constraints)
type LogEntry
    Level string
    Message string

entries := logs
    |> slice.GroupBy(func(e LogEntry) string {
        return e.Level
    })
# Result: map[string][]LogEntry with keys like "ERROR", "WARN", "INFO"

# Available functions:
# First, Last, Drop, DropLast, Reverse
# Unique, Chunk, Filter, Map, GroupBy
```

**Note on GroupBy:** This function uses Go 1.25+ generics with proper type constraints:
- `GroupBy[T any, K comparable](items []T, keyFunc func(T) K) map[K][]T`
- The `K` type parameter is constrained to `comparable` (required for map keys)

### String Package

String operations designed for pipes:

```kukicha
import "stdlib/string"

# Text processing pipeline
result := rawText
    |> string.TrimSpace()
    |> string.ToLower()
    |> string.ReplaceAll("_", "-")
    |> string.Split("\n")
    |> slice.Filter(func(line string) bool {
        return not string.IsEmpty(line)
    })

# URL cleanup
cleanUrl := url
    |> string.TrimPrefix("https://")
    |> string.TrimSuffix("/")
    |> string.Split("/")
    |> slice.Last(1)

# Available functions:
# ToUpper, ToLower, TrimSpace, Trim, TrimPrefix, TrimSuffix
# Split, Join, Contains, HasPrefix, HasSuffix, Replace, ReplaceAll
```

### Concurrent Package

Concurrency helpers leveraging Go 1.25+ `sync.WaitGroup.Go()`:

```kukicha
import "stdlib/concurrent"

# Run multiple tasks concurrently
concurrent.Parallel(
    func() {
        fetchUsers()
    },
    func() {
        fetchOrders()
    },
    func() {
        fetchProducts()
    }
)

# Run with concurrency limit (max 4 at a time)
tasks := list of func(){}
for url in urls
    tasks = append(tasks, func() {
        processUrl(url)
    })

concurrent.ParallelWithLimit(4, tasks...)

# Track a goroutine with WaitGroup
wg := concurrent.Go(func() {
    processLargeFile()
})
# Do other work...
wg.Wait()

# Available functions:
# Parallel, ParallelWithLimit, Go
```

### HTTP Package

HTTP server helpers using Go 1.25+ features:

```kukicha
import "stdlib/http"
import "net/http"

func main()
    mux := http.NewServeMux()
    mux.HandleFunc("/api/data", handleData)

    # Wrap with CSRF protection (Go 1.25+ CrossOriginProtection)
    handler := mux |> http.WithCSRF()

    # Start server
    http.Serve(":8080", handler) onerr panic "server failed"

func handleData(w http.ResponseWriter, r reference http.Request)
    data := fetchData()
    json.NewEncoder(w).Encode(data)

# Available functions:
# WithCSRF, Serve
```

### Fetch Package 

HTTP client with fluent builder pattern and response helpers. **Important**: JSON parsing uses Go 1.25+ jsonv2 directly for type safety.

**request building:**
```kukicha
import "stdlib/fetch"

# Builder pattern with headers and timeouts
resp, err := fetch.New("https://api.example.com/data")
    |> fetch.Header("Authorization", "Bearer token")
    |> fetch.Timeout(30 * time.Second)
    |> fetch.Do()

# Simple GET request
resp, err := fetch.Get("https://api.github.com/users")

# POST request
resp, err := fetch.Post(data, "https://api.example.com/users")

# Multiple headers and custom timeout
resp, err := fetch.New("https://api.example.com/data")
    |> fetch.Header("Authorization", "Bearer token")
    |> fetch.Header("Content-Type", "application/json")
    |> fetch.Timeout(60 * time.Second)
    |> fetch.Method("POST")
    |> fetch.Do()
```

**response parsing:**
```kukicha
import "stdlib/fetch"
import "stdlib/json"
import "stdlib/slice"

# Example 1: Simple text response
text := fetch.Get("https://api.example.com/version")
    |> fetch.CheckStatus()
    |> fetch.Text()
    onerr panic "fetch failed"

# Example 2: Typed JSON with stdlib/json - Simple approach
type User
    ID int json:"id"
    Name string json:"name"
    Followers int json:"followers"

user := User{}
fetch.Get("https://api.github.com/users/golang")
    |> fetch.CheckStatus()
    |> fetch.Bytes()
    |> json.Unmarshal(reference user)
    onerr panic "fetch failed"

print("User: {user.Name} with {user.Followers} followers")

# Example 3: Streaming JSON (for large responses)
type Repo
    Name string json:"name"
    Stars int json:"stargazers_count"
    Archived bool json:"archived"

resp, err := fetch.Get("https://api.github.com/users/golang/repos")
if err != empty
    panic("fetch failed")
defer resp.Body.Close()

repos := list of Repo{}
resp.Body |> json.NewDecoder() |> .Decode(reference repos) onerr panic

# Now filter with slice helpers - beautiful pipes!
active := repos
    |> slice.Filter(func(r Repo) bool {
        return not r.Archived and r.Stars > 100
    })

print("Found {len(active)} active repos")

# Example 4: POST with auto-serialization (uses stdlib/json internally)
newUser := User{Name: "Alice"}
resp := fetch.Post(newUser, "https://api.example.com/users")
    |> fetch.CheckStatus()
    onerr panic "create failed"
```

**Design Philosophy:**

- Use `fetch.Bytes()` + `json.Unmarshal()` for simple cases
- Use streaming with `json.NewDecoder()` for large responses
- `fetch.Post()` auto-serializes request body using stdlib/json
- No `fetch.Json()` helper - Go's type system requires knowing the target type at compile time, so we provide `Bytes()` for use with `stdlib/json` instead

### JSON Package 

Pipe-friendly wrapper around Go 1.25+ jsonv2 for beautiful syntax with 2-10x performance.

```kukicha
import "stdlib/json"
import "net/http"

type Todo
    ID int json:"id"
    Title string json:"title"
    Completed bool json:"completed"

# Simple encoding with pipes
func sendTodo(w http.ResponseWriter, r reference http.Request)
    todo := Todo{ID: 1, Title: "Learn Kukicha", Completed: false}

    w.Header().Set("Content-Type", "application/json")
    w |> json.NewEncoder() |> .Encode(todo) onerr return

# Pretty-printed JSON with builder pattern
func sendPretty(w http.ResponseWriter, r reference http.Request)
    data := MyData{...}

    w
        |> json.NewEncoder()
        |> json.WithDeterministic()
        |> json.WithIndent("  ")
        |> .Encode(data)
        onerr return

# Decoding from request
func createTodo(w http.ResponseWriter, r reference http.Request)
    todo := Todo{}

    r.Body
        |> json.NewDecoder()
        |> .Decode(reference todo)
        onerr return w.WriteHeader(400)

    # Use the todo...
    print("Created: {todo.Title}")

# Convenience functions for simple cases
jsonBytes := json.Marshal(data) onerr panic
prettyJson := json.MarshalPretty(config) onerr panic
json.Unmarshal(jsonBytes, reference result) onerr panic
```

### Parse Package 

Universal parsing for CSV and YAML. **For JSON parsing, use `stdlib/json` directly** (see JSON package above).

```kukicha
import "stdlib/parse"
import "stdlib/json"
import "stdlib/files"

# Define struct with JSON tags
type Config
    Port int json:"port"
    Host string json:"host"
    Debug bool json:"debug"

# JSON parsing: Use stdlib/json directly for type safety
config := Config{}
"config.json"
    |> files.Read()
    |> json.Unmarshal(reference config)
    onerr panic "parse failed"

# Or with pipes using json.NewDecoder:
file := files.Open("config.json") onerr panic
defer file.Close()
config := Config{}
file |> json.NewDecoder() |> .Decode(reference config) onerr panic

# Format as pretty JSON (convenience function)
output := config
    |> parse.JsonPretty()
    |> files.Write("config-formatted.json")

# CSV to structured data
users := "data.csv"
    |> files.Read()
    |> parse.Csv()
    |> slice.Drop(1)              # Skip header
    |> slice.Map(csvRowToUser)
    |> slice.Filter(func(u User) bool {
        return u.Active
    })

# YAML config parsing (requires Go interop)
# Note: Use gopkg.in/yaml.v3 directly for YAML parsing

# YAML formatting (convenience function)
settings := config
    |> parse.YamlPretty()
    |> files.Write("config.yaml")
```

**Struct Tags:** JSON and other parsers use struct tags for automatic field mapping:
```kukicha
type User
    ID int json:"id"
    Name string json:"name"
    Email string json:"email"
```

### CLI Package 

Build command-line tools easily:

```kukicha
import "stdlib/cli"

func main()
    # Parse command line arguments
    flags, positional := cli.Parse()
    cmd := cli.Command(positional)
    
    if cmd == ""
        cli.PrintUsage("mytool", ["fetch", "process"])
        return
    
    # Get arguments and flags
    input := cli.String(flags, positional, "1")
    verbose := cli.BoolFlag(flags, "verbose")
    format := cli.Flag(flags, "format")
    
    if cmd == "fetch"
        # Process fetch command
        url := cli.String(flags, positional, "1")
        print "Fetching {url} with format {format}"
    else if cmd == "process"
        # Process process command
        output := cli.String(flags, positional, "2")
        print "Processing {input} to {output}"
```

**Key Features:**
- Simple argument parsing with `cli.Parse()`
- Command extraction with `cli.Command()`
- Flag handling with `cli.Flag()`, `cli.BoolFlag()`, `cli.IntFlag()`
- Positional argument access with `cli.String()`
- Usage help with `cli.PrintUsage()`

### Files Package 

File operations optimized for pipes.

```kukicha
import "stdlib/files"

# Read and process file
output := "input.txt"
    |> files.Read()
    |> string.Split("\n")
    |> slice.Filter(func(line string) bool {
        return not string.IsEmpty(line)
    })
    |> slice.Map(string.TrimSpace)
    |> slice.Map(processLine)
    |> string.Join("\n")
    |> files.Write("output.txt")
    onerr panic "processing failed"

# Check if file exists
if files.Exists("config.yaml")
    loadConfig()
else
    createDefaultConfig()

# List files with filtering
logs := files.List("/var/log")
    |> slice.Filter(func(f string) bool {
        return string.HasSuffix(f, ".log")
    })
    |> slice.Map(func(f string) string {
        return f
    })

# Watch for changes (useful for dev tools)
files.Watch("./src/**/*.kuki", func(path string) {
    print("Changed: {path}")
    rebuildProject()
})

# Temp file handling with automatic cleanup
files.TempFile("test-") |> files.UseWith(func(path string) {
    files.Write(path, data) onerr panic "write failed"
    processFile(path)
})
```

### Shell Package

Safe command execution without shell injection with builder pattern support.

```kukicha
import "stdlib/shell"

# Run a simple command
output := shell.RunSimple("ls", "-la") onerr panic "ls failed"
print output

# Builder pattern: Run with options
cmd := shell.New("git", "status") |> shell.Dir("./repo") |> shell.SetTimeout(30)
result := shell.Execute(cmd)

if shell.Success(result)
    print(string(shell.GetOutput(result)))
else
    print("Error: {string(shell.GetError(result))}")

# Direct execution (legacy)
output := shell.RunSimple("pwd") onerr panic "command failed"
print(string(output))

# Check if command exists
if shell.Which("docker")
    print "Docker is installed"
else
    print "Docker not found"

# Run with environment variables
cmd := shell.New("npm", "install") |> shell.Dir("./frontend") |> shell.Env("NODE_ENV", "production")
result := shell.Execute(cmd)
if shell.Success(result)
    print("npm install succeeded")
```

**Key Features:**
- Builder pattern with `shell.New()`, `shell.Dir()`, `shell.SetTimeout()`, `shell.Env()`
- Command execution with `shell.Execute()` returning `Result` type
- Result inspection with `shell.Success()`, `shell.GetOutput()`, `shell.GetError()`, `shell.ExitCode()`
- Legacy direct functions: `shell.Run()`, `shell.RunSimple()`, `shell.RunWithDir()`
- Command existence checking with `shell.Which()`
- Environment variable helpers: `shell.Getenv()`, `shell.Setenv()`

## Additional Scripting Packages

### Template Package ✅

Text templating for code generation and reports:

```kukicha
import "stdlib/template"

# Simple string templating
email := template.Render("Hello {{.Name}}, your order #{{.OrderId}} is ready!")
    |> template.Data(map of string to any{
        "Name": user.Name,
        "OrderId": order.Id,
    })
    onerr panic "template failed"

# File-based templates
report := "report.tmpl"
    |> files.Read()
    |> template.Parse()
    |> template.Data(reportData)
    |> template.Render()
    |> files.Write("report.html")

# Code generation
code := template.New()
    |> template.Parse(codeTemplate)
    |> template.Data(structDef)
    |> template.Render()
    |> files.Write("generated.go")
```

### Retry Package ⚠️

⚠️ **Status: Limited Implementation** - The retry package provides helper functions for manual retry patterns. Automatic retry with `retry.Do()` requires language features not yet supported. See stdlib/retry/retry.kuki for working examples.

Manual retry pattern with helper functions:

```kukicha
import "stdlib/retry"
import "stdlib/fetch"

# Manual retry with configuration
func fetchWithRetry(url string) (list of byte, error)
    cfg := retry.New()
        |> retry.Attempts(5)
        |> retry.Delay(500)
        |> retry.Backoff(1)  # 1 = Exponential

    attempt := 0
    for attempt < cfg.MaxAttempts
        # Try the operation
        data := fetch.Get(url)
            |> fetch.CheckStatus()
            |> fetch.Bytes()
            onerr discard

        # Check if it succeeded
        if data != empty
            return data, empty

        # Sleep with exponential backoff before next attempt
        if attempt < cfg.MaxAttempts - 1
            retry.Sleep(cfg, attempt)

        attempt = attempt + 1

    return empty, error("all retries failed")

# Simple retry with linear backoff
func processWithRetry() bool
    cfg := retry.New()
        |> retry.Attempts(3)
        |> retry.Delay(1000)
        |> retry.Backoff(0)  # 0 = Linear

    attempt := 0
    for attempt < cfg.MaxAttempts
        success := processData() onerr discard
        if success
            return true

        retry.Sleep(cfg, attempt)
        attempt = attempt + 1

    return false
```

**Note:** The automatic `retry.Do()` pattern shown in many retry libraries requires passing functions that return `(value, error)` tuples, which conflicts with Kukicha's `onerr` operator. The manual pattern above is the recommended approach.

### Result Package ✅

✅ **Status: Implemented** - Optional and Result types for educational purposes and explicit error handling patterns.

Optional and Result types:

```kukicha
import "stdlib/result"

# Optional type (better than empty/nil for beginners)
user := findUserById(123)  # Returns Optional[User]

if user.IsSome()
    print "Found: {user.Unwrap().Name}"
else
    print "User not found"

# Or use pipeline
message := findUserById(123)
    |> result.Map(func(u User) string {
        return "Hello, {u.Name}"
    })
    |> result.UnwrapOr("User not found")

# Result type (explicit error handling)
data := parseConfig("config.yaml")  # Returns Result[Config, Error]

configResult := data
    |> result.MapErr(func(e error) error {
        return error("Config error: {e}")
    })

config := configResult
    |> result.Unwrap()
    onerr defaultConfig()

# Chaining operations
output := result.Ok(initialData)
    |> result.AndThen(validate)
    |> result.AndThen(transform)
    |> result.AndThen(save)

# Check result
if output.IsOk()
    data := output.Unwrap()
    message := "Success: saved {data.Id}"
else
    err := output.Err()
    message := "Failed: {err}"
```

**Type Assertions:** Kukicha supports type assertions using the `as` keyword with multi-value assignment:
```kukicha
inner, ok := opt.value as Optional
if ok
    # Type assertion succeeded, inner is of type Optional
```

---

## Real-World Scripting Examples

These examples show how Kukicha excels at practical automation.

### Example 1: API Data Processing Script

```kukicha
import "stdlib/fetch"
import "stdlib/json"
import "stdlib/slice"
import "stdlib/files"

type Repo
    Name string json:"name"
    Stars int json:"stargazers_count"
    Archived bool json:"archived"
    HtmlUrl string json:"html_url"

type RepoSummary
    Name string
    Stars int
    Url string

func main()
    # Fetch and parse repos from GitHub API (simple approach)
    repos := list of Repo{}
    fetch.Get("https://api.github.com/users/golang/repos")
        |> fetch.CheckStatus()
        |> fetch.Bytes()
        |> json.Unmarshal(reference repos)
        onerr panic "failed to fetch repos"

    # Filter and transform using slice helpers
    summaries := repos
        |> slice.Filter(func(r Repo) bool {
            return not r.Archived and r.Stars > 100
        })
        |> slice.Map(func(r Repo) RepoSummary {
            return RepoSummary{
                Name: r.Name,
                Stars: r.Stars,
                Url: r.HtmlUrl,
            }
        })

    # Save to file using json.MarshalPretty
    output := summaries
        |> json.MarshalPretty()
        onerr panic "failed to serialize"

    string(output)
        |> files.Write("active-repos.json")
        onerr panic "failed to save"

    print("Saved {len(summaries)} active repositories")
```

### Example 2: Log Analysis Tool

```kukicha
import "stdlib/files"
import "stdlib/cli"

func main()
    app := cli.New("logparse") |> cli.Arg("logfile", "Path to log file") |> cli.AddFlag("level", "Filter by level (ERROR|WARN|INFO)", "ERROR") |> cli.Action(analyzeLog)
    cli.RunApp(app) onerr panic "command failed"

func analyzeLog(args cli.Args)
    logPath := cli.GetString(args, "logfile")
    level := cli.GetString(args, "level")

    errors := logPath
        |> files.Read()
        |> string.Split("\n")
        |> slice.Filter(func(line string) bool {
            return string.Contains(line, level)
        })
        |> slice.Map(parseLine)
        |> slice.GroupBy(func(e LogEntry) string {
            return e.ErrorCode
        })
        |> summarize()

    print "Found {len(errors)} {level} entries"
    errors |> printSummary()
```

### Example 3: File Processing Pipeline

```kukicha
import "stdlib/files"
import "stdlib/template"

# Convert CSV to HTML report
func main()
    report := "sales.csv"
        |> files.Read()
        |> parse.Csv()
        |> slice.Drop(1)              # Remove header
        |> slice.Map(csvToSale)
        |> calculateTotals()
        |> template.Render(reportTemplate)
        |> files.Write("report.html")
        onerr panic "report generation failed"

    print "Report saved to report.html"
```

### Example 4: Deployment Automation

```kukicha
import "stdlib/shell"
import "stdlib/files"

func deploy(version string)
    print "Building version {version}..."

    # Build and test
    shell.Run("go", "test", "./...")
        |> shell.Dir("./backend")
        |> shell.Must()

    shell.Run("npm", "run", "build")
        |> shell.Dir("./frontend")
        |> shell.Must()

    # Create deployment package
    tmpDir := files.TempDir() onerr panic "temp dir failed"
    shell.Run("cp", "-r", "dist", tmpDir)
    shell.Run("tar", "-czf", "deploy-{version}.tar.gz", tmpDir)
    # Note: Manual cleanup needed - useWith() helper not yet implemented

    print "Deployment package ready: deploy-{version}.tar.gz"
```

---

## Using Go Standard Library Directly

You can ALWAYS use Go packages directly with pipes and `onerr`:

### HTTP with Go stdlib

```kukicha
import "net/http"
import "encoding/json"
import "io"

# GET request pipeline
users := http.Get("https://api.example.com/users")
    |> checkStatus()
    |> .Body
    |> io.ReadAll()
    |> json.Unmarshal(_, reference users) as list of User
    onerr empty list of User

func checkStatus(resp reference http.Response, err error) (reference http.Response, error)
    if err != empty
        return empty, err
    if resp.StatusCode >= 400
        return empty, error "request failed: {resp.Status}"
    return resp, empty
```

### File I/O with Go stdlib

```kukicha
import "os"
import "bufio"

# Read lines and process
output := "input.txt"
    |> os.Open()
    |> readLines()
    |> slice.Filter(isValid)
    |> slice.Map(transform)
    |> string.Join("\n")
    |> writeFile("output.txt")
    onerr panic "processing failed"

func readLines(file reference os.File) list of string
    defer file.Close()
    lines := list of string{}
    scanner := bufio.NewScanner(file)
    for scanner.Scan()
        lines = append(lines, scanner.Text())
    return lines
```

---

## Design Rationale

### Why Scripting Packages?

1. **Real Pain Point**: Go is verbose for one-off scripts
2. **Pipe-First Design**: Every function works naturally in pipelines
3. **Beginner-Friendly**: Hide common patterns (JSON parsing, HTTP, etc.)
4. **Still Just Go**: All packages use Go stdlib underneath
5. **Educational**: Great for learning programming concepts

### Why Not Wrap Everything?

We provide high-level helpers for common patterns, but:

- ✅ Provide: Helpers Go lacks (fetch.CheckStatus/Bytes/Text, files.Read piping)
- ✅ Provide: Ergonomic patterns (retry, CLI parsing)
- ❌ Don't wrap: Every Go stdlib function (maintenance hell)
- ❌ Don't duplicate: Functionality Go already does well

### Comparison to Other Languages

**vs Python:**
- ✅ Kukicha: Compiled binary, Go speed, type safety
- ✅ Python: Larger ecosystem, no compilation step

**vs Bash:**
- ✅ Kukicha: Type safe, readable, maintainable
- ✅ Bash: Installed everywhere, ultimate compatibility

**vs Go:**
- ✅ Kukicha: Less verbose for scripts, pipes, better error handling
- ✅ Go: More explicit, larger community, job market

---

## Contributing

We welcome contributions! Focus areas:

### High Priority:
- **Implementing scripting packages** 
- **Writing examples** showcasing pipe-based workflows
- **Tutorial content** for beginners learning programming

### Good Contributions:
- Improving existing packages (iter, slice, string)
- Documentation with real-world examples
- Performance optimizations
- Bug fixes

### Guidelines:
1. **Every function should work in pipes** - First parameter is data to transform
2. **Error handling with `onerr`** - Return tuples, let users handle errors
3. **Build on Go stdlib** - Don't reinvent wheels, provide convenience
4. **Focus on scripting** - Use cases: automation, data processing, CLI tools
5. **Beginner-friendly** - Clear names, good docs, simple APIs
