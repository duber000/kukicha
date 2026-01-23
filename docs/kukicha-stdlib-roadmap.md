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

### ✅ Completed Packages

| Package | Purpose | Key Feature |
|---------|---------|-------------|
| **iter** | Functional iteration (Filter, Map, Reduce) | Lazy evaluation with pipes |
| **slice** | Slice operations (First, Last, Drop, Unique) | Pipeline-friendly helpers |
| **string** | String utilities | Thin wrappers for Go's strings package |

### 🚧 Planned Scripting Packages (Priority Order)

These packages make Kukicha perfect for scripts and automation:

| Package | Purpose | Status |
|---------|---------|--------|
| **fetch** | HTTP client optimized for pipes | Planned |
| **parse** | JSON/YAML/CSV/TOML/XML parsing | Planned |
| **cli** | CLI argument parsing made easy | Planned |
| **files** | File operations with pipes | Planned |
| **shell** | Safe command execution | Planned |
| **template** | Text templating | Planned |
| **retry** | Retry logic with backoff | Planned |
| **result** | Optional/Result types (educational) | Planned |

---

## Completed Packages ✅

### Iter Package

Functional iteration with lazy evaluation and pipes:

```kukicha
import "stdlib/iter"

# Pipeline: filter positive numbers, double them, sum
total := numbers
    |> iter.Filter(n -> n > 0)
    |> iter.Map(n -> n * 2)
    |> iter.Reduce(0, (acc, n) -> acc + n)

# Find first matching item
user := users
    |> iter.Find(u -> u.Email equals "admin@example.com")
    |> unwrapOr(createDefaultUser())

# Take first 10, skip first 2
page := items
    |> iter.Skip(2)
    |> iter.Take(10)
    |> iter.Collect()

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
    |> slice.Map(u -> u.ID)
    |> slice.First(10)

# Batch processing
batches := items
    |> slice.Chunk(100)
    |> slice.Map(processBatch)

# Available functions:
# First, Last, Drop, DropLast, Reverse
# Unique, Chunk, Filter, Map
```

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
    |> slice.Filter(line -> not string.IsEmpty(line))

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

---

## Planned Scripting Packages 🚧

These packages showcase the pipe operator and make scripting delightful:

### Fetch Package (Planned)

HTTP client designed for data pipelines:

```kukicha
import "stdlib/fetch"

# Simple GET with automatic JSON parsing
users := fetch.Get("https://api.github.com/users")
    |> fetch.Json() as list of User
    |> slice.Filter(u -> u.Followers > 100)
    |> slice.Map(u -> u.Login)
    onerr empty list of string

# POST with JSON body
response := user
    |> fetch.JsonBody()
    |> fetch.Post("https://api.example.com/users")
    |> fetch.Text()
    onerr "request failed"

# Pipeline with headers and timeouts
data := fetch.New("https://api.example.com/data")
    |> fetch.Header("Authorization", "Bearer {token}")
    |> fetch.Timeout(30.seconds)
    |> fetch.Get()
    |> fetch.Json() as Response
    onerr panic "fetch failed"

# Parallel fetching
results := urls
    |> slice.Map(url -> fetch.Get(url) |> fetch.Text())
    |> waitAll()
```

### Parse Package (Planned)

Universal parsing with pipes:

```kukicha
import "stdlib/parse"

# JSON parsing pipeline
config := "config.json"
    |> files.Read()
    |> parse.Json() as Config
    |> validateConfig()
    onerr defaultConfig()

# CSV to structured data
users := "data.csv"
    |> files.Read()
    |> parse.Csv()
    |> slice.Drop(1)              # Skip header
    |> slice.Map(csvRowToUser)
    |> slice.Filter(u -> u.Active)

# YAML config with validation
settings := "settings.yaml"
    |> files.Read()
    |> parse.Yaml() as Settings
    |> applyDefaults()
    onerr panic "invalid settings"

# TOML (like stem.toml)
project := "stem.toml"
    |> files.Read()
    |> parse.Toml() as ProjectConfig

# XML parsing
feed := fetch.Get("https://example.com/rss")
    |> parse.Xml() as RssFeed
    |> extractItems()
```

### CLI Package (Planned)

Build command-line tools easily:

```kukicha
import "stdlib/cli"

func main()
    app := cli.New("mytool")
        |> cli.Description("A tool for processing data")
        |> cli.Version("1.0.0")

    app.Command("fetch")
        |> cli.Arg("url", "URL to fetch")
        |> cli.Flag("format", "Output format (json|text)", "json")
        |> cli.Action(fetchCommand)

    app.Command("process")
        |> cli.Arg("input", "Input file")
        |> cli.Flag("output", "Output file", "output.txt")
        |> cli.Action(processCommand)

    app.Run() onerr panic "command failed"

func fetchCommand(args cli.Args)
    url := args.String("url")
    format := args.String("format")

    data := fetch.Get(url)
        |> fetch.Text()
        onerr panic "fetch failed"

    if format equals "json"
        data |> parse.Json() |> print
    else
        print data
```

### Files Package (Planned)

File operations optimized for pipes:

```kukicha
import "stdlib/files"

# Read and process file
output := "input.txt"
    |> files.Read()
    |> string.Split("\n")
    |> slice.Filter(line -> not string.IsEmpty(line))
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
    |> slice.Filter(f -> string.HasSuffix(f.Name, ".log"))
    |> slice.Filter(f -> f.ModTime.After(yesterday))
    |> slice.Map(f -> f.Path)

# Watch for changes (useful for dev tools)
files.Watch("./src/**/*.kuki", func(path string)
    print "Changed: {path}"
    rebuildProject()
)

# Temp file handling
result := files.TempFile()
    |> useWith(tempFile ->
        tempFile |> files.Write(data)
        processFile(tempFile.Path)
    )  # Auto-deleted after use
```

### Shell Package (Planned)

Safe command execution without shell injection:

```kukicha
import "stdlib/shell"

# Run command and capture output
output := shell.Run("git", "status")
    |> shell.Output()
    |> string.TrimSpace()
    onerr "git failed"

# Pipeline commands (safer than bash pipes)
result := shell.Run("cat", "data.txt")
    |> shell.Pipe(shell.Run("grep", "ERROR"))
    |> shell.Pipe(shell.Run("wc", "-l"))
    |> shell.Output()

# Run with timeout and working directory
exitCode := shell.New("npm", "install")
    |> shell.Dir("./frontend")
    |> shell.Timeout(5.minutes)
    |> shell.Env("NODE_ENV", "production")
    |> shell.Run()
    |> shell.ExitCode()
    onerr panic "npm install failed"

# Stream output in real-time
shell.Run("docker", "build", ".")
    |> shell.Stdout(line -> print "[build] {line}")
    |> shell.Stderr(line -> print "[error] {line}")
    |> shell.Wait()
    onerr panic "build failed"

# Check if command exists
if shell.Which("docker")
    runDockerBuild()
else
    print "Docker not installed"
```

### Template Package (Planned)

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
    |> template.AddFunc("title", string.ToTitle)
    |> template.Parse(codeTemplate)
    |> template.Data(structDef)
    |> template.Render()
    |> files.Write("generated.go")
```

### Retry Package (Planned)

Retry logic with exponential backoff:

```kukicha
import "stdlib/retry"

# Retry with default backoff (3 attempts)
data := retry.Do(func() any {
        return fetch.Get(apiUrl) |> fetch.Json()
    })
    onerr panic "all retries failed"

# Custom retry strategy
result := retry.New()
    |> retry.Attempts(5)
    |> retry.Delay(1.second)
    |> retry.Backoff(retry.Exponential)
    |> retry.Do(func() any {
        return processData()
    })
    onerr logError("processing failed after 5 attempts")

# Retry with condition (only retry on specific errors)
response := retry.DoIf(
    func() any { return callExternalApi() },
    func(err error) bool { return isRetryable(err) }
)
```

### Result Package (Planned)

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
    |> result.Map(u -> "Hello, {u.Name}")
    |> result.UnwrapOr("User not found")

# Result type (explicit error handling)
data := parseConfig("config.yaml")  # Returns Result[Config, Error]

config := data
    |> result.MapErr(e -> "Config error: {e}")
    |> result.Unwrap()
    onerr defaultConfig()

# Chaining operations
output := result.Ok(initialData)
    |> result.AndThen(validate)
    |> result.AndThen(transform)
    |> result.AndThen(save)
    |> result.Match(
        ok: d -> "Success: saved {d.Id}",
        err: e -> "Failed: {e}"
    )
```

---

## Real-World Scripting Examples

These examples show how Kukicha excels at practical automation:

### Example 1: API Data Processing Script

```kukicha
import "stdlib/fetch"
import "stdlib/parse"
import "stdlib/files"

# Fetch GitHub repos, filter active ones, save to JSON
func main()
    repos := fetch.Get("https://api.github.com/users/golang/repos")
        |> fetch.Json() as list of Repo
        |> slice.Filter(r -> not r.Archived)
        |> slice.Filter(r -> r.Stars > 100)
        |> slice.Map(r -> RepoSummary{
            Name: r.Name,
            Stars: r.Stars,
            Url: r.HtmlUrl,
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
        |> slice.Filter(line -> string.Contains(line, level))
        |> slice.Map(parseLine)
        |> slice.GroupBy(e -> e.ErrorCode)
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

## Implementation Priorities

To make Kukicha the best scripting language for beginners and automation:

### Phase 1: Essential Scripting (Next)
1. **fetch** - HTTP client (most requested, enables tons of examples)
2. **parse** - JSON/YAML/CSV parsing (pairs with fetch)
3. **files** - File operations (basic scripting need)

### Phase 2: Tools & CLI (After Phase 1)
4. **cli** - Argument parsing (build actual tools)
5. **shell** - Safe command execution (automation scripts)
6. **template** - Text generation (code gen, reports)

### Phase 3: Advanced (Future)
7. **retry** - Reliability patterns
8. **result** - Educational type for learning FP concepts

---

## Contributing

We welcome contributions! Focus areas:

### High Priority:
- **Implementing scripting packages** (fetch, parse, files, cli, shell)
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

---

## Success Metrics

We'll know Kukicha stdlib is successful when:

- ✅ Complete beginners can fetch API data and save to JSON in 10 lines
- ✅ Go developers reach for Kukicha for one-off scripts
- ✅ "Script it in Kukicha" blog posts appear
- ✅ Educational content uses Kukicha for teaching
- ✅ Someone says "I love the pipe operator" unprompted

---

**Last Updated:** 2026-01-23
**Philosophy Document:** [kukicha-design-philosophy.md](kukicha-design-philosophy.md)
**Target Users:** Beginners learning programming, Go developers writing scripts, automation enthusiasts
