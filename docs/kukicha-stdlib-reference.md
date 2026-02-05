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
- Kukicha can call any Go stdlib directly without wrappers
- We also provide convenient wrapper packages for common scripting patterns
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

**Ready to Use:** ✅ iterator, slice, string, files, json, parse (CSV/YAML), fetch, concurrent, shell, cli, http, template, result, validate, must, env, datetime

**Partially Implemented:** ⚠️ retry (limited stub due to error handling design constraints)

**Limitations:**
- Constants with types not supported
- Retry package has limited functionality - see package documentation

### Completed Packages

| Package | Purpose | Status | Functions |
|---------|---------|--------|-----------|
| **iterator** | Functional iteration (Filter, Map, Reduce) | ✅ Ready | Filter, Map, FlatMap, Take, Skip, Reduce, Collect, Find, Any, All, Enumerate, Zip, Chunk |
| **slice** | Slice operations with generics | ✅ Ready | First, Last, Drop, DropLast, Reverse, Unique, Chunk, Filter, Map, Contains, IndexOf, Concat, GroupBy, Get, GetOr, FirstOne, FirstOr, LastOne, LastOr, Find, FindOr, FindIndex, FindLast, IsEmpty, IsNotEmpty, Pop, Shift |
| **string** | String utilities | ✅ Ready | ToUpper, ToLower, Title, Trim, TrimSpace, TrimPrefix, TrimSuffix, Split, Join, Contains, HasPrefix, HasSuffix, Index, Count, Replace, ReplaceAll, and more |
| **files** | File operations with pipes | ✅ Ready | Read, Write, Append, Exists, IsDir, IsFile, List, Delete, Copy, Move, MkDir, TempFile, TempDir, Size, ModTime, Extension, Join, Abs, Watch, UseWith |
| **json** | Pipe-friendly jsonv2 wrapper | ✅ Ready | NewEncoder, WithDeterministic, WithIndent, Encode, NewDecoder, Decode, Marshal, MarshalPretty, Unmarshal, MarshalWrite, UnmarshalRead |
| **parse** | CSV/YAML parsing | ✅ Ready | JsonPretty, Csv, CsvWithHeader, YamlPretty |
| **fetch** | HTTP client with json integration | ✅ Ready | Get, Post, New, Header, Timeout, Method, Body, Do, CheckStatus, Text, Bytes |
| **concurrent** | Concurrency helpers | ✅ Ready | Parallel, ParallelWithLimit, Go |
| **shell** | Safe command execution | ✅ Ready | Builder: New, Dir, SetTimeout, Env, Execute. Result helpers: Success, GetOutput, GetError, ExitCode. Utilities: Which, Getenv, Setenv, Environ |
| **cli** | CLI argument parsing | ✅ Ready | Builder: New, Arg, AddFlag, Action, RunApp. Args helpers: GetString, GetBool, GetInt |
| **http** | HTTP server helpers | ✅ Ready | WithCSRF, Serve, JSON, JSONStatus, JSONError, JSONBadRequest, JSONNotFound, ReadJSON, GetQueryParam, GetQueryInt, GetHeader, Text, HTML, IsGet, IsPost, MethodNotAllowed |
| **template** | Text templating | ✅ Ready | Render, Data, Execute, Parse, New, WithContent, RenderSimple, Must |
| **result** | Optional/Result types | ✅ Ready | Some, None, Ok, Err, Map, UnwrapOr, AndThen, Match, ToOptional, FromOptional, Flatten, FlattenResult, All, Any |
| **validate** | Input validation helpers | ✅ Ready | NotEmpty, MinLength, MaxLength, LengthBetween, Matches, Email, URL, Alpha, Alphanumeric, Numeric, StartsWith, EndsWith, Contains, OneOf, Positive, Negative, InRange, Min, Max, ParseInt, ParseFloat, ParseBool, NotEmptyList |
| **must** | Initialization helpers (panic on error) | ✅ Ready | Do, DoMsg, Ok, OkMsg, Env, EnvOr, EnvInt, EnvIntOr, EnvBool, EnvBoolOr, EnvList, EnvListOr, True, False, NotEmpty, NotNil |
| **env** | Typed environment variable access | ✅ Ready | Get, GetOr, GetInt, GetIntOr, GetIntOrDefault, GetBool, GetBoolOr, GetBoolOrDefault, GetFloat, GetFloatOr, GetList, GetListOr, Set, Unset, IsSet, IsSetAndNotEmpty, All |
| **datetime** | Time helpers with named formats | ✅ Ready | Format, Parse, Now, Today, Tomorrow, Yesterday, Seconds, Minutes, Hours, Days, Weeks, AddDays, SubDays, IsBefore, IsAfter, IsBetween, IsToday, IsPast, IsFuture, Unix, FromUnix, Sleep, InUTC, InLocal |

### Partially Implemented

| Package | Purpose | Status | Notes |
|---------|---------|--------|-------|
| **retry** | Full retry logic with automatic error handling | ⚠️ Partial | Manual retry helpers available (recommended approach). Automatic retry.Do() not available yet due to language design constraints. See stdlib/retry/retry.kuki for examples. |

### Known Limitations

Some roadmap examples use aspirational syntax not yet supported:

| Feature | What's Not Supported | Workaround |
|---------|---------------------|------------|
| **retry** | Automatic retry with `retry.Do()` | Use manual retry loops (see stdlib/retry/retry.kuki) |
| **shell** | Command piping with `shell.Pipe()` | Use shell pipes directly or multiple commands |

### Error Handling Best Practices

When using `onerr`, follow these guidelines:

```kukicha
# ✅ Use panic only for unrecoverable startup errors
config := loadConfig() onerr
    panic "missing config file"

# ✅ Return errors from library functions
function ProcessFile(path string) (data any, error)
    return path |> files.Read() onerr
        return empty, error

# ✅ Log and continue for recoverable errors in scripts
data := fetch.Get(url) onerr
    log.Printf("Warning: failed to fetch {url}")
    return empty  # or provide a default

# ❌ Don't use panic in production handlers
result := operation() onerr
    panic "should not happen"  # BAD!
```

**Rule of Thumb:**
- Use `panic` for **startup errors** (missing config, bad flags)
- Use `return error` for **library functions**
- Use `onerr log.Printf()` for **recoverable script errors**

---

### Iterator Package

Functional iteration with lazy evaluation and pipes:

```kukicha
import "stdlib/iterator"
import "stdlib/slice"

# Pipeline: filter positive numbers, double them, sum
# Note: iterator provides functional composition; slice provides eager operations
total := numbers
    |> slice.Filter(function(n int) bool
        return n > 0  # Keep only positive numbers
    )
    |> slice.Map(function(n int) int
        return n * 2  # Double each number
    )
    |> iterator.Reduce(0, function(acc int, n int) int
        return acc + n  # Sum all numbers
    )

# Available slice operations:
# Filter, Map, Drop, DropLast, First, Last
# Unique, Chunk, Reverse, Contains, GroupBy

# Available iterator operations (functional composition):
# Filter, Map, FlatMap, Take, Skip, Enumerate, Zip
# Chunk, Reduce, Collect, Any, All, Find

# DevOps Example: Analyzing high-latency requests from logs
longRequests := logLines
    |> iterator.Map(parseLogLine)
    |> iterator.Filter(func(e LogEntry) bool
        return e.Duration > 5 * time.Second
    )
    |> iterator.Take(10)
    |> iterator.Collect()
```

**Note:** `slice` operations work on entire collections eagerly, while `iterator` provides lazy functional composition. Use `slice` for direct transformations and `iterator` for complex composed workflows.

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
    |> slice.Map(function(u User) int
        return u.ID
    )
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
    |> slice.GroupBy(function(e LogEntry) string
        return e.Level
    )
# Result: map[string][]LogEntry with keys like "ERROR", "WARN", "INFO"

# Available functions:
# First, Last, Drop, DropLast, Reverse
# Unique, Chunk, Filter, Map, GroupBy

# DevOps Example: Filtering healthy nodes
healthyNodes := nodes
    |> slice.Filter(func(n Node) bool
        return n.Status == "Ready" and n.CPUUsage < 80.0
    )
    |> slice.Map(func(n Node) string
        return n.ID
    )
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
    |> slice.Filter(func(line string) bool
        return not string.IsEmpty(line)
    )

# URL cleanup
cleanUrl := url
    |> string.TrimPrefix("https://")
    |> string.TrimSuffix("/")
    |> string.Split("/")
    |> slice.Last(1)

# Available functions:
# ToUpper, ToLower, TrimSpace, Trim, TrimPrefix, TrimSuffix
# Split, Join, Contains, HasPrefix, HasSuffix, Replace, ReplaceAll

# DevOps Example: Parsing resource ARN
resourceName := "arn:aws:s3:::my-bucket-name"
    |> string.Split(":")
    |> slice.Last(1)
    |> string.Join("") # or just use slice.Last(1) and indexing
```

### Concurrent Package

Concurrency helpers leveraging Go 1.25+ `sync.WaitGroup.Go()`:

```kukicha
import "stdlib/concurrent"

# Run multiple tasks concurrently
concurrent.Parallel(
    func()
        fetchUsers()
    ,
    func()
        fetchOrders()
    ,
    func()
        fetchProducts()
)

# Run with concurrency limit (max 4 at a time)
tasks := list of func(){}
for url in urls
    tasks = append(tasks, func()
        processUrl(url)
    )

concurrent.ParallelWithLimit(4, tasks...)

# Run a function in a goroutine
concurrent.Go(func()
    processLargeFile()
)

# Available functions:
# Parallel, ParallelWithLimit, Go

# DevOps Example: Pinging multiple heartbeat endpoints concurrently
urls := list of string{"https://api1.status.com", "https://api2.status.com"}
tasks := list of func(){}
for u in urls
    url := u
    tasks = append(tasks, func()
        fetch.Get(url) |> fetch.CheckStatus() onerr print "fail: {url}"
    )
concurrent.Parallel(tasks...)
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

# DevOps Example: Service health check endpoint
mux.HandleFunc("/health", func(w http.ResponseWriter, r reference http.Request)
    w.WriteHeader(200)
    w.Write(list of byte("OK"))
)
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
    onerr
        panic "fetch failed"

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
    onerr
        panic "fetch failed"

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
resp.Body |> json.NewDecoder() |> json.Decode(_, reference repos) onerr panic

# Now filter with slice helpers - beautiful pipes!
active := repos
    |> slice.Filter(func(r Repo) bool
        return not r.Archived and r.Stars > 100
    )

print("Found {len(active)} active repos")

# Example 4: POST with auto-serialization (uses stdlib/json internally)
newUser := User
    Name: "Alice"
resp := fetch.Post(newUser, "https://api.example.com/users")
    |> fetch.CheckStatus()
    onerr
    panic "create failed"

# DevOps Example: Calling AWS MetaData Service or K8s API
token := fetch.Get("http://169.254.169.254/latest/api/token")
    |> fetch.Header("X-aws-ec2-metadata-token-ttl-seconds", "21600")
    |> fetch.Method("PUT")
    |> fetch.Text()
    onerr ""

instanceId := fetch.Get("http://169.254.169.254/latest/meta-data/instance-id")
    |> fetch.Header("X-aws-ec2-metadata-token", token)
    |> fetch.Text()
    onerr "unknown"
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
function sendTodo(w http.ResponseWriter, r reference http.Request)
    todo := Todo
        ID: 1
        Title: "Learn Kukicha"
        Completed: false

    w.Header().Set("Content-Type", "application/json")
    w |> json.NewEncoder() |> .Encode(todo) onerr
        return

# Pretty-printed JSON with builder pattern
function sendPretty(w http.ResponseWriter, r reference http.Request)
    data := MyData{...}

    w
        |> json.NewEncoder()
        |> json.WithDeterministic()
        |> json.WithIndent("  ")
        |> .Encode(data)
        onerr
            return

# Decoding from request
func createTodo(w http.ResponseWriter, r reference http.Request)
    todo := Todo{}

    r.Body
        |> json.NewDecoder()
        |> json.Decode(_, reference todo)
        onerr return w.WriteHeader(400)

    # Use the todo...
    print("Created: {todo.Title}")

# Convenience functions for simple cases
jsonBytes := json.Marshal(data) onerr panic
json.Unmarshal(jsonBytes, reference result) onerr panic

# DevOps Example: Parsing infrastructure config
config := DatabaseConfig{}
"config.json"
    |> files.Read()
    |> json.Unmarshal(reference config)
    onerr panic "invalid config"
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
file |> json.NewDecoder() |> json.Decode(_, reference config) onerr panic

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
    |> slice.Filter(func(u User) bool
        return u.Active
    )

# YAML config parsing (requires Go interop)
# Note: Use gopkg.in/yaml.v3 directly for YAML parsing

# YAML formatting (convenience function)
settings := config
    |> parse.YamlPretty()
    |> files.Write("config.yaml")

# DevOps Example: Converting K8s JSON to YAML
"pod.json"
    |> files.Read()
    |> json.Unmarshal(reference pod)
    |> parse.YamlPretty()
    |> files.Write("pod.yaml")
```

**Struct Tags:** JSON and other parsers use struct tags for automatic field mapping:
```kukicha
type User
    ID int json:"id"
    Name string json:"name"
    Email string json:"email"
```

### CLI Package

Build command-line tools easily with builder pattern:

```kukicha
import "stdlib/cli"

func main()
    app := cli.New("mytool")
        |> cli.Arg("command", "Command to run (fetch|process)")
        |> cli.Arg("input", "Input file or URL")
        |> cli.AddFlag("verbose", "Enable verbose output", "false")
        |> cli.AddFlag("format", "Output format", "json")
        |> cli.Action(handleCommand)

    cli.RunApp(app) onerr panic "command failed"

func handleCommand(args cli.Args)
    command := cli.GetString(args, "command")
    input := cli.GetString(args, "input")
    verbose := cli.GetBool(args, "verbose")
    format := cli.GetString(args, "format")

    if command == "fetch"
        print("Fetching {input} with format {format}")
    else if command == "process"
        output := cli.GetString(args, "output")
        print("Processing {input} to {output}")

# DevOps Example: Deployment CLI flag
app := cli.New("deployer")
    |> cli.AddFlag("env", "Target environment (prod|staging)", "staging")
    |> cli.AddFlag("dry-run", "Show changes without applying", "true")
    |> cli.Action(doDeploy)
```

**Key Features:**
- Builder pattern with `cli.New()`, `cli.Arg()`, `cli.AddFlag()`, `cli.Action()`
- Application execution with `cli.RunApp()`
- Argument access with `cli.GetString()`, `cli.GetBool()`, `cli.GetInt()`
- Clean separation of app definition and handler logic

### Files Package 

File operations optimized for pipes.

```kukicha
import "stdlib/files"

# Read and process file
output := "input.txt"
    |> files.Read()
    |> string.Split("\n")
    |> slice.Filter(func(line string) bool
        return not string.IsEmpty(line)
    )
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
    |> slice.Filter(func(f string) bool
        return string.HasSuffix(f, ".log")
    )
    |> slice.Map(func(f string) string
        return f
    )

# Watch for changes (useful for dev tools)
files.Watch("./src/**/*.kuki", func(path string)
    print("Changed: {path}")
    rebuildProject()
)

# Temp file handling with automatic cleanup
files.TempFile("test-") |> files.UseWith(func(path string)
    files.Write(path, data) onerr panic "write failed"
    processFile(path)
)

# DevOps Example: Cleaning up old logs
files.List("/var/log/app")
    |> slice.Filter(func(f string) bool
        return strings.HasSuffix(f, ".gz") and files.ModTime(f) |> datetime.IsPast()
    )
    |> slice.Map(files.Delete)
```

### Shell Package

Safe command execution without shell injection (commands bypass shell parsing).

⚠️ **Security Note:** `shell.New` executes the command directly without shell parsing. While this prevents shell injection, always validate and sanitize user input before passing it as command arguments. Do not pass unsanitized user input as arguments.

```kukicha
import "stdlib/shell"

# Simple command
cmd := shell.New("ls", "-la")
result := shell.Execute(cmd)
if shell.Success(result)
    print(string(shell.GetOutput(result)))

# Run with options using builder pattern
cmd := shell.New("git", "status") |> shell.Dir("./repo") |> shell.SetTimeout(30)
result := shell.Execute(cmd)

if shell.Success(result)
    print(string(shell.GetOutput(result)))
else
    print("Error: {string(shell.GetError(result))}")

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

# ⚠️ SAFE: Build args programmatically
cmd := shell.New("grep", searchTerm, filename)  # Direct args, no shell parsing

# ⚠️ UNSAFE: Don't pass unsanitized user input
userQuery := getUserInput()  # Could be malicious
cmd := shell.New("grep", userQuery, file)  # DANGEROUS if userQuery isn't validated

# DevOps Example: Safe service restart
if shell.Success(shell.New("systemctl", "restart", "nginx") |> shell.Execute())
    print "Nginx restarted"
```

**Key Features:**
- Builder pattern with `shell.New()`, `shell.Dir()`, `shell.SetTimeout()`, `shell.Env()`
- Command execution with `shell.Execute()` returning `Result` type
- Result inspection with `shell.Success()`, `shell.GetOutput()`, `shell.GetError()`, `shell.ExitCode()`
- Command existence checking with `shell.Which()`
- Environment variable helpers: `shell.Getenv()`, `shell.Setenv()`, `shell.Environ()`

### Validate Package

Input validation helpers designed for pipes and composability.

```kukicha
import "stdlib/validate"

# DevOps Example: Validating infrastructure config
config := map of string to string{
    "region": "us-east-1",
    "nodes": "3",
    "email": "admin@example.com",
}

# Validation pipeline
region := config["region"] 
    |> validate.NotEmpty() 
    |> validate.OneOf("us-east-1", "us-west-2", "eu-west-1")
    onerr panic "invalid region"

nodeCount := config["nodes"]
    |> validate.ParsePositiveInt()
    |> validate.Max(10)
    onerr panic "nodes must be 1-10"

adminEmail := config["email"]
    |> validate.Email()
    onerr panic "invalid admin email"

# Available functions:
# NotEmpty, MinLength, MaxLength, Length, LengthBetween, Matches, Email, URL
# Alpha, Alphanumeric, Numeric, NoWhitespace, StartsWith, EndsWith, Contains, OneOf
# Positive, Negative, NonNegative, NonZero, InRange, Min, Max
# ParseInt, ParsePositiveInt, ParseFloat, ParseBool
# NotEmptyList, ListMinLength, ListMaxLength
```

### Must Package

Initialization helpers that panic on error. Use these ONLY for startup/initialization code where failure should stop the process immediately (fail-fast).

```kukicha
import "stdlib/must"

# DevOps Example: CI/CD runner initialization
# Fail fast if required environment variables or tools are missing
func init()
    # Ensure required secrets are set
    apiKey := must.Env("DEPLOY_KEY")
    dbPass := must.Env("DB_PASSWORD")

    # Ensure required tools are installed
    must.Ok(shell.Which("docker") |> boolToError("docker not found"))
    must.Ok(shell.Which("kubectl") |> boolToError("kubectl not found"))

    # Load config or fail
    config := must.Do(loadConfig("config.yaml"))
    
    print "Runner initialized successfully"

# Available functions:
# Do, DoMsg, Ok, OkMsg
# Env, EnvOr, EnvInt, EnvIntOr, EnvBool, EnvBoolOr, EnvList, EnvListOr
# True, False, NotEmpty, NotNil
```

## Additional Scripting Packages

### Template Package

Text templating for code generation and reports:

```kukicha
import "stdlib/template"

# Simple string templating
email := template.Render("Hello {{.Name}}, your order #{{.OrderId}} is ready!")
    |> template.Data(map of string to any{
        "Name": user.Name,
        "OrderId": order.Id,
    })
    onerr
        panic "template failed"

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

# DevOps Example: Generating a Dockerfile
dockerfile := template.Render(dockerTemplate)
    |> template.Data(map of string to string{"BaseImage": "alpine:latest", "Port": "8080"})
    |> files.Write("Dockerfile")
```

### Retry Package

**Status: Partial Implementation** - The retry package provides helper functions for manual retry patterns. Automatic retry with `retry.Do()` requires language features not yet supported. See stdlib/retry/retry.kuki for working examples.

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

# DevOps Example: Retrying a database migration
func migrateWithRetry()
    cfg := retry.New() |> retry.Attempts(10) |> retry.Delay(2000)
    attempt := 0
    for attempt < cfg.MaxAttempts
        shell.New("migrate", "-up") |> shell.Execute() onerr
            print "Migration failed, retrying..."
            retry.Sleep(cfg, attempt)
            attempt = attempt + 1
            continue
        print "Migration successful"
        return
```

**Note:** The automatic `retry.Do()` pattern shown in many retry libraries requires passing functions that return `(value, error)` tuples, which conflicts with Kukicha's `onerr` operator. The manual pattern above is the recommended approach.

### Result Package

**Status: Implemented** - Optional and Result types for educational purposes and explicit error handling patterns.

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
    |> result.Map(func(u User) string
        return "Hello, {u.Name}"
    )
    |> result.UnwrapOr("User not found")

# Result type (explicit error handling)
data := parseConfig("config.yaml")  # Returns Result[Config, Error]

configResult := data
    |> result.MapErr(func(e error) error
        return error("Config error: {e}")
    )

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

# DevOps Example: Optional deployment metadata
func getDeploymentTag() result.Optional[string]
    tag := env.Get("DEPLOY_TAG") onerr return result.None[string]()
    return result.Some(tag)
```

**Type Assertions:** Kukicha supports type assertions using the `as` keyword with multi-value assignment:
```kukicha
inner, ok := opt.value as Optional
if ok
    # Type assertion succeeded, inner is of type Optional

### Env Package

Typed access to environment variables with safe defaults and `onerr` support.

```kukicha
import "stdlib/env"

# DevOps Example: Reading service configuration
func loadConfig() Config
    return Config
        Port: env.GetIntOrDefault("PORT", 8080)
        Debug: env.GetBoolOrDefault("DEBUG", false)
        LogFormat: env.GetOr("LOG_FORMAT", "json")
        AllowedHosts: env.GetListOr("ALLOWED_HOSTS", ",", list of string{"localhost"})
        ApiKey: env.Get("API_KEY") onerr "temporary-key"

# Check if environment is prepared
if env.IsSetAndNotEmpty("KUBERNETES_SERVICE_HOST")
    print "Running in Kubernetes"

# Available functions:
# Get, GetOr, GetInt, GetIntOr, GetIntOrDefault, GetBool, GetBoolOr, GetBoolOrDefault
# GetFloat, GetFloatOr, GetList, GetListOr, Set, Unset, IsSet, IsSetAndNotEmpty, All
```

### DateTime Package

Human-friendly time formatting and duration helpers.

```kukicha
import "stdlib/datetime"

# DevOps Example: Checking resource expiration
func checkExpiration(createdAt time.Time)
    # Use named durations
    ninetyDays := datetime.Days(90)
    expirationDate := createdAt |> datetime.AddDays(90)

    if datetime.Now() |> datetime.IsAfter(expirationDate)
        print "Resource expired on {datetime.Format(expirationDate, "date")}"
        cleanupResource()

# Formatting examples
now := datetime.Now()
print "ISO8601: {datetime.Format(now, "iso8601")}"
print "Date: {datetime.Format(now, "date")}"
print "Kitchen: {datetime.Format(now, "kitchen")}"

# Available functions:
# Format, Parse, ParseInLocation, Now, Today, Tomorrow, Yesterday
# Nanoseconds, Microseconds, Milliseconds, Seconds, Minutes, Hours, Days, Weeks
# AddDays, AddWeeks, AddMonths, AddYears, SubDays, SubWeeks, SubMonths, SubYears
# IsBefore, IsAfter, IsBetween, IsSameDay, IsToday, IsYesterday, IsTomorrow, IsPast, IsFuture
# Year, Month, Day, Hour, Minute, Second, Weekday, WeekdayName
# Unix, UnixMilli, FromUnix, FromUnixMilli, Sleep, SleepSeconds, SleepMilliseconds
# InUTC, InLocal, InLocation
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
        |> slice.Filter(function(r Repo) bool
            return not r.Archived and r.Stars > 100
        )
        |> slice.Map(function(r Repo) RepoSummary
            return RepoSummary
                Name: r.Name
                Stars: r.Stars
                Url: r.HtmlUrl
        )

    # Save to file using json.MarshalPretty
    output := summaries
        |> json.MarshalPretty()
        onerr
            panic "failed to serialize"

    string(output)
        |> files.Write("active-repos.json")
        onerr
            panic "failed to save"

    print("Saved {len(summaries)} active repositories")
```

### Example 2: Log Analysis Tool

```kukicha
import "stdlib/files"
import "stdlib/cli"

function main()
    app := cli.New("logparse")
        |> cli.Arg("logfile", "Path to log file")
        |> cli.AddFlag("level", "Filter by level (ERROR|WARN|INFO)", "ERROR")
        |> cli.Action(analyzeLog)

    cli.RunApp(app) onerr
        panic "command failed"

function analyzeLog(args cli.Args)
    logPath := cli.GetString(args, "logfile")
    level := cli.GetString(args, "level")

    errors := logPath
        |> files.Read()
        |> string.Split("\n")
        |> slice.Filter(func(line string) bool
            return string.Contains(line, level)
        )
        |> slice.Map(parseLine)
        |> slice.GroupBy(func(e LogEntry) string
            return e.ErrorCode
        )
        |> summarize()

    print "Found {len(errors)} {level} entries"
    errors |> printSummary()
```

### Example 3: File Processing Pipeline

```kukicha
import "stdlib/files"
import "stdlib/template"

# Convert CSV to HTML report
function main()
    report := "sales.csv"
        |> files.Read()
        |> parse.Csv()
        |> slice.Drop(1)              # Remove header
        |> slice.Map(csvToSale)
        |> calculateTotals()
        |> template.Render(reportTemplate)
        |> files.Write("report.html")
        onerr
            panic "report generation failed"

    print "Report saved to report.html"
```

### Example 4: Deployment Automation

```kukicha
import "stdlib/shell"
import "stdlib/files"

function deploy(version string)
    print("Building version {version}...")

    # Build and test
    result := shell.New("go", "test", "./...") |> shell.Dir("./backend") |> shell.Execute()
    if not shell.Success(result)
        panic("tests failed: {string(shell.GetError(result))}")

    result2 := shell.New("npm", "run", "build") |> shell.Dir("./frontend") |> shell.Execute()
    if not shell.Success(result2)
        panic("build failed: {string(shell.GetError(result2))}")

    # Create deployment package
    tmpDir := files.TempDir() onerr
        panic "temp dir failed"
    shell.New("cp", "-r", "dist", tmpDir) |> shell.Execute()
    shell.New("tar", "-czf", "deploy-{version}.tar.gz", tmpDir) |> shell.Execute()

    print("Deployment package ready: deploy-{version}.tar.gz")
```

---

## DevOps & SRE Examples

Kukicha is exceptionally well-suited for infrastructure automation. Below are advanced patterns combining multiple standard library packages.

### Automated SSL Certificate Check

This script checks a list of domains for SSL expiration and sends an alert if any are expiring within 30 days.

```kukicha
import "stdlib/fetch"
import "stdlib/datetime"
import "stdlib/concurrent"
import "stdlib/env"

func main()
    domains := env.GetListOr("DOMAINS", ",", list of string{"example.com", "google.com"})
    webhookUrl := must.Env("SLACK_WEBHOOK")

    tasks := list of func(){}
    for d in domains
        domain := d
        tasks = append(tasks, func()
            checkDomain(domain, webhookUrl)
        )
    
    concurrent.Parallel(tasks...)

func checkDomain(domain string, webhook string)
    # Note: Using native Go tls package via interop
    import "crypto/tls"
    
    conn, err := tls.Dial("tcp", "{domain}:443", empty)
    if err != empty
        return print "Failed to connect to {domain}: {err}"
    defer conn.Close()

    # Get expiration from certificate
    expiry := conn.ConnectionState().PeerCertificates[0].NotAfter
    
    # Check if expiring in 30 days
    if datetime.Now() |> datetime.AddDays(30) |> datetime.IsAfter(expiry)
        msg := "Alert: Certificate for {domain} expires on {datetime.Format(expiry, "date")}"
        fetch.Post(msg, webhook) onerr print "Failed to send alert"
```

### Infrastructure Drift Detector

Compares current local configuration files against a remote state API.

```kukicha
import "stdlib/files"
import "stdlib/fetch"
import "stdlib/json"
import "stdlib/slice"

type State
    ResourceId string json:"id"
    Version    string json:"version"

func main()
    # 1. Fetch remote state
    remoteState := list of State{}
    fetch.Get("https://api.infra.local/state")
        |> fetch.CheckStatus()
        |> fetch.Bytes()
        |> json.Unmarshal(reference remoteState)
        onerr panic "could not fetch state"

    # 2. Read local state
    localState := "infra-state.json"
        |> files.Read()
        |> json.Unmarshal(reference localState)
        onerr list of State{}

    # 3. Find missing or changed resources
    drift := slice.Filter(remoteState, func(r State) bool
        found := slice.Find(localState, func(l State) bool
            return l.ResourceId == r.ResourceId
        )
        return not found.IsSome() or found.Unwrap().Version != r.Version
    )

    if len(drift) > 0
        print "Infrastructure drift detected in {len(drift)} resources!"
    else
        print "Infrastructure is in sync."
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
    onerr
        empty list of User

function checkStatus(resp reference http.Response, err error) (reference http.Response, error)
    if err != empty
        return empty, err
    if resp.StatusCode >= 400
        return empty, error("request failed: {resp.Status}")
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
    onerr
        panic "processing failed"

function readLines(file reference os.File) list of string
    defer file.Close()
    variable lines = list of string{}
    scanner := bufio.NewScanner(file)
    for scanner.Scan()
        lines = append(lines, scanner.Text())
    return lines
```

---

## Design Rationale

### Why Scripting Packages?

We provide scripting packages because:

- **Real Pain Point** - Go is verbose for one-off scripts
- **Pipe-First Design** - Every function works naturally in pipelines
- **Beginner-Friendly** - Hide common patterns (JSON parsing, HTTP, etc.)
- **Still Just Go** - All packages use Go stdlib underneath
- **Educational** - Great for learning programming concepts

### Why Not Wrap Everything?

We provide high-level helpers for common patterns, but:

- **Do provide:** Helpers Go lacks (fetch.CheckStatus/Bytes/Text, files.Read piping)
- **Do provide:** Ergonomic patterns (retry, CLI parsing)
- **Don't wrap:** Every Go stdlib function (maintenance nightmare)
- **Don't duplicate:** Functionality Go already does well

---

## Contributing

We welcome contributions! Focus areas:

### High Priority:
- **Implementing scripting packages** 
- **Writing examples** showcasing pipe-based workflows
- **Tutorial content** for beginners learning programming

### Good Contributions:
- Improving existing packages (iterator, slice, string)
- Documentation with real-world examples
- Performance optimizations
- Bug fixes

### Guidelines:
1. **Every function should work in pipes** - First parameter is data to transform
2. **Error handling with `onerr`** - Return tuples, let users handle errors
3. **Build on Go stdlib** - Don't reinvent wheels, provide convenience
4. **Focus on scripting** - Use cases: automation, data processing, CLI tools
5. **Beginner-friendly** - Clear names, good docs, simple APIs
