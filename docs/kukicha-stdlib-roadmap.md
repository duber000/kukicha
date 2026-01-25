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

**Ready to Use:** ✅ iter, slice, string, files, parse, fetch, concurrent, shell, cli, http (basic)

**Not Yet Implemented:** 🚧 template, retry, result packages

**Limitations:**
- Builder patterns for `shell.New()`, `cli.New()` not yet supported - use direct function calls
- `fetch.New()` builder pattern is now ✅ implemented
- `files.Watch()`, `useWith()` helper not implemented
- Some Go 1.25+ features in roadmap examples are aspirational

### ✅ Completed Packages

| Package | Purpose | Status | Functions |
|---------|---------|--------|-----------|
| **iter** | Functional iteration (Filter, Map, Reduce) | ✅ Ready | Filter, Map, FlatMap, Take, Skip, Reduce, Collect, Find, Any, All, Enumerate, Zip, Chunk |
| **slice** | Slice operations with generics | ✅ Ready | First, Last, Drop, DropLast, Reverse, Unique, Chunk, Filter, Map, Contains, IndexOf, Concat, GroupBy |
| **string** | String utilities | ✅ Ready | ToUpper, ToLower, Title, Trim, TrimSpace, TrimPrefix, TrimSuffix, Split, Join, Contains, HasPrefix, HasSuffix, Index, Count, Replace, ReplaceAll, and more |
| **files** | File operations with pipes | ✅ Ready | Read, Write, Append, Exists, IsDir, IsFile, List, Delete, Copy, Move, MkDir, TempFile, TempDir, Size, ModTime, Extension, Join, Abs |
| **parse** | JSON/YAML/CSV parsing | ✅ Ready | Json, JsonLines, JsonPretty, Csv, CsvWithHeader, Yaml, YamlPretty |
| **fetch** | HTTP client optimized for pipes | ✅ Ready | Get, Post, Json, Text, CheckStatus |
| **concurrent** | Concurrency helpers | ✅ Ready | Parallel, ParallelWithLimit, Go |
| **shell** | Safe command execution | ✅ Ready | Run, RunSimple, RunWithDir, RunWithOutput, RunWithTimeout, Which, Getenv, Setenv, Unsetenv, Environ |
| **cli** | CLI argument parsing | ✅ Ready | Parse, String, Command, Flag, BoolFlag, IntFlag, PrintUsage |
| **http** | HTTP server helpers | ⚠️ Limited | Basic helpers (no builder pattern yet) |

### 🚧 Planned Scripting Packages (Priority Order)

These packages make Kukicha perfect for scripts and automation:

| Package | Purpose | Status | Notes |
|---------|---------|--------|-------|
| **template** | Text templating | 🚧 Not Started | Planned but not implemented. Roadmap examples won't work yet. |
| **retry** | Retry logic with backoff | 🚧 Not Started | Planned but not implemented. Roadmap examples won't work yet. |
| **result** | Optional/Result types (educational) | 🚧 Not Started | Planned but not implemented. Roadmap examples won't work yet. |

### ⚠️ Partially Implemented (Roadmap Examples May Not Work)

| Feature | Status | What Works | What Doesn't |
|---------|--------|-----------|--------------|
| **fetch** | ✅ Complete | `Get()`, `Post()`, `New()`, `Header()`, `Timeout()`, `Method()`, `Do()` | `Json()`, `Text()`, `CheckStatus()` helpers require Go interop |
| **shell** | Mostly works | `Run()`, `RunSimple()`, direct execution | Builder pattern (`shell.New().Dir()`) not implemented |
| **cli** | Mostly works | Simple parsing | Builder pattern (`cli.New().Arg()`) not implemented |
| **files** | Mostly works | Basic file operations | `Watch()` and `useWith()` helper not implemented |

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

---

## Planned Scripting Packages 🚧

These packages showcase the pipe operator and make scripting delightful:

### Fetch Package ⚠️

HTTP client with fluent builder pattern. Request building complete, response parsing in progress.

**Partially Implemented:**
- ✅ `New(url)` - Create a request builder
- ✅ `Header(req, name, value)` - Add HTTP header (chainable)
- ✅ `Timeout(req, duration)` - Set timeout (chainable)
- ✅ `Method(req, method)` - Set HTTP method (chainable)
- ✅ `Do(req)` - Execute the request
- ✅ `Get(url)` - Quick GET request
- ✅ `Post(data, url)` - Quick POST request
- 🚧 `Json(resp)` - Parse JSON with streaming support
- 🚧 `Text(resp)` - Read response as text
- 🚧 `CheckStatus(resp)` - Verify HTTP status code (2xx)

**Working Examples (request building):**
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

**Aspirational Examples (need response parsing helpers):**
```kukicha
# These examples will work once Json(), Text(), CheckStatus() are implemented

# Parse JSON with streaming
data := fetch.Get("https://api.github.com/users")
    |> fetch.CheckStatus()
    |> fetch.Json() as list of User
    |> slice.Filter(func(u User) bool {
        return u.Followers > 100
    })

# Read response as text
text := fetch.New("https://api.example.com/data")
    |> fetch.Header("Authorization", "Bearer token")
    |> fetch.Do()
    |> fetch.Text()
```

### Parse Package ✅

Universal parsing with pipes (Go 1.25+ jsonv2 for 2-10x faster JSON). Works seamlessly with struct tags for automatic field mapping.

```kukicha
import "stdlib/parse"

# Define struct with JSON tags
type Config
    Port int json:"port"
    Host string json:"host"
    Debug bool json:"debug"

# JSON parsing pipeline (uses Go 1.25+ jsonv2)
config := "config.json"
    |> files.Read()
    |> parse.Json() as Config
    |> validateConfig()
    onerr defaultConfig()

# Stream JSON from reader (memory efficient for large files)
data := fileReader
    |> parse.JsonFromReader() as LargeDataset
    onerr panic "failed to parse"

# Parse NDJSON (newline-delimited JSON logs)
entries := logData
    |> parse.JsonLines() as list of LogEntry
    |> slice.Filter(func(e LogEntry) bool {
        return e.Level equals "ERROR"
    })

# Format as pretty JSON
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

# YAML config with validation
settings := "settings.yaml"
    |> files.Read()
    |> parse.Yaml() as Settings
    |> applyDefaults()
    onerr panic "invalid settings"
```

**Struct Tags:** JSON and other parsers use struct tags for field mapping:
```kukicha
type User
    ID int json:"id"
    Name string json:"name"
    Email string json:"email"
```

# Available functions:
# Json, JsonFromReader, JsonLines, JsonPretty (Go 1.25+ jsonv2)
# Csv, CsvWithHeader, Yaml
```

### CLI Package ✅

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

**Available functions:** Parse, Command, String, Flag, BoolFlag, IntFlag, PrintUsage

### Files Package ✅

File operations optimized for pipes.

⚠️ **Note:** `files.Watch()` and the `useWith()` helper below are not yet implemented - use other functions like `Read()`, `Write()`, `List()`, `TempFile()`, and `TempDir()` which work great with pipes.

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
    |> slice.Filter(func(f FileInfo) bool {
        return string.HasSuffix(f.Name, ".log")
    })
    |> slice.Filter(func(f FileInfo) bool {
        return f.ModTime.After(yesterday)
    })
    |> slice.Map(func(f FileInfo) string {
        return f.Path
    })

# Watch for changes (useful for dev tools)
files.Watch("./src/**/*.kuki", func(path string) {
    print("Changed: {path}")
    rebuildProject()
})

# Temp file handling
result := files.TempFile()
    |> useWith(tempFile ->
        tempFile |> files.Write(data)
        processFile(tempFile.Path)
    )  # Auto-deleted after use
```

### Shell Package ✅

Safe command execution without shell injection.

⚠️ **Note:** Builder pattern examples below (`shell.New().Dir().Timeout().Run()`) and piping (`shell.Pipe()`) are aspirational - use `Run()`, `RunSimple()`, and `RunWithDir()` for now.

```kukicha
import "stdlib/shell"

# Run a simple command
output := shell.RunSimple("ls", "-la") onerr panic "ls failed"
print output

# Run with options
result := shell.New("git", "status")
    |> shell.Dir("./repo")
    |> shell.Timeout(30 * time.Second)
    |> shell.Run()

if shell.Success(result)
    print result.stdout
else
    print "Error: {result.stderr}"

# Pipeline commands
count := shell.New("cat", "data.txt")
    |> shell.Pipe(shell.New("grep", "ERROR"))
    |> shell.Pipe(shell.New("wc", "-l"))
    |> shell.Run()
    |> shell.Output()

# Check if command exists
if shell.Which("docker")
    print "Docker is installed"
else
    print "Docker not found"

# Run with environment variables
output := shell.New("npm", "install")
    |> shell.Dir("./frontend")
    |> shell.Env("NODE_ENV", "production")
    |> shell.Run()
    |> shell.Output()
```

**Key Features:**
- Safe command execution with `shell.Run()` and `shell.RunSimple()`
- Command piping with `shell.Pipe()`
- Environment and directory control with `shell.Env()` and `shell.Dir()`
- Timeout support with `shell.Timeout()`
- Result inspection with `shell.Output()`, `shell.Error()`, `shell.ExitCode()`, `shell.Success()`
- Command existence checking with `shell.Which()`

**Available functions:** New, Dir, Env, Timeout, Run, RunSimple, Output, Error, ExitCode, Success, Which, Pipe

## Planned Scripting Packages 🚧

### Template Package (Planned) 🚧

🚧 **Status: Not Implemented** - This package does not exist yet. The examples below are aspirational and won't currently work.

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

### Retry Package (Planned) 🚧

🚧 **Status: Not Implemented** - This package does not exist yet. The examples below are aspirational and won't currently work.

Retry logic with exponential backoff:

```kukicha
import "stdlib/retry"

# Retry with default backoff (3 attempts)
data := retry.Do(func() any {
    return fetch.Get(apiUrl) |> fetch.Json()
}) onerr panic "all retries failed"

# Custom retry strategy
result := retry.New()
    |> retry.Attempts(5)
    |> retry.Delay(1000)        # milliseconds
    |> retry.Backoff(retry.Exponential)
    |> retry.Do(func() any {
        return processData()
    })
    onerr logError("processing failed after 5 attempts")

# Retry with condition (only retry on specific errors)
response := retry.DoIf(
    func() any {
        return callExternalApi()
    },
    func(err error) bool {
        return isRetryable(err)
    }
)
```

### Result Package (Planned) 🚧

🚧 **Status: Not Implemented** - This package does not exist yet. The examples below are aspirational and won't currently work.

Optional and Result types for educational purposes:

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

---

## Real-World Scripting Examples

These examples show how Kukicha excels at practical automation.

**Implementation Status:**
- ✅ **Example 1** (API Data Processing) - **Works!** Uses fetch, parse, slice, files - all implemented
- ✅ **Example 2** (Log Analysis) - **Works!** Uses files, slice.GroupBy, parse, string - all implemented
- ✅ **Example 3** (File Processing) - **Works!** Uses files, parse, slice, string - all implemented
- ✅ **Example 4** (Deployment) - **Works!** Uses shell, files, basic command execution - all implemented

### Example 1: API Data Processing Script

```kukicha
import "stdlib/fetch"
import "stdlib/parse"
import "stdlib/files"

# Fetch GitHub repos, filter active ones, save to JSON
func main()
    repos := fetch.Get("https://api.github.com/users/golang/repos")
        |> fetch.Json() as list of Repo
        |> slice.Filter(func(r Repo) bool {
            return not r.Archived
        })
        |> slice.Filter(func(r Repo) bool {
            return r.Stars > 100
        })
        |> slice.Map(func(r Repo) RepoSummary {
            return RepoSummary{
                Name: r.Name,
                Stars: r.Stars,
                Url: r.HtmlUrl,
            }
        })
        |> parse.ToJson()
        |> files.Write("active-repos.json")
        onerr panic "failed to process repos"

    print "Saved {len(repos)} active repositories"
```

### Example 2: Log Analysis Tool

```kukicha
import "stdlib/files"
import "stdlib/cli"

func main()
    app := cli.New("logparse")
        |> cli.Arg("logfile", "Path to log file")
        |> cli.Flag("level", "Filter by level (ERROR|WARN|INFO)", "ERROR")
        |> cli.Action(analyzeLog)

    app.Run() onerr panic "command failed"

func analyzeLog(args cli.Args)
    logPath := args.String("logfile")
    level := args.String("level")

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
    files.TempDir()
        |> useWith(tmpDir ->
            shell.Run("cp", "-r", "dist", tmpDir)
            shell.Run("tar", "-czf", "deploy-{version}.tar.gz", tmpDir)
        )

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
    |> parse.Json() as list of User
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

- ✅ Provide: Helpers Go lacks (fetch.Json, files.Read piping)
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
