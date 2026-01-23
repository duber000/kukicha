# Kukicha Standard Library

The Kukicha standard library provides **value-add packages** that extend Go's capabilities. Following the CoffeeScript model ("It's just Go"), we only include packages that provide functionality Go lacks or makes awkward.

## Philosophy

Kukicha doesn't wrap Go's standard library - you use it directly with `onerr` syntax:

```kukicha
import "encoding/json"
import "net/http"

# Use Go stdlib directly
data := json.Marshal(user) onerr return error
resp := http.Get(url) onerr return nil, error
```

The Kukicha stdlib only exists where it adds genuine value.

## Packages

### iter - Functional Iterator Operations

Go's `iter.Seq` protocol is low-level. We provide higher-level functional operations:

```kukicha
import "slices"
import "stdlib/iter"

numbers := list of int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

# Lazy pipeline
result := slices.Values(numbers)
    |> iter.Filter(func(n int) bool { return n > 3 })
    |> iter.Map(func(n int) int { return n * 2 })
    |> iter.Take(5)
    |> iter.Collect()

# result = [8, 10, 12, 14, 16]
```

**Functions:** Filter, Map, FlatMap, Take, Skip, Enumerate, Zip, Chunk, Reduce, Collect, Any, All, Find

### slice - Slice Operations

Common slice operations Go lacks:

```kukicha
import "stdlib/slice"

firstThree := slice.First(items, 3)
lastTwo := slice.Last(items, 2)
reversed := slice.Reverse(items)
unique := slice.Unique(items)
chunked := slice.Chunk(items, 5)
```

**Functions:** First, Last, Drop, DropLast, Reverse, Unique, Chunk, Filter, Map, Contains, IndexOf, Concat

### string - String Utilities

Thin wrappers with minimal maintenance burden:

```kukicha
import "stdlib/string"

upper := string.ToUpper("hello")
trimmed := string.TrimSpace("  hello  ")
parts := string.Split("a,b,c", ",")
```

**Functions:** ToUpper, ToLower, TrimSpace, TrimPrefix, TrimSuffix, Split, Join, Fields, Lines, Contains, HasPrefix, HasSuffix, ReplaceAll, EqualFold

### fetch - HTTP Client for Pipes

HTTP client designed for pipeline-based data fetching:

```kukicha
import "stdlib/fetch"

# Simple GET and JSON parsing
users := fetch.Get("https://api.example.com/users")
    |> fetch.CheckStatus()
    |> fetch.Json() as list of User
    onerr empty list of User

# POST with JSON body
response := userData
    |> fetch.Post("https://api.example.com/users")
    |> fetch.Text()
    onerr "request failed"

# Advanced request builder
data := fetch.New("https://api.example.com/data")
    |> fetch.Header("Authorization", "Bearer {token}")
    |> fetch.Timeout(30 * time.Second)
    |> fetch.Do()
    |> fetch.Json() as Response
```

**Functions:** Get, Post, Json, Text, CheckStatus, Status, New, Header, Timeout, Body, Do, PostRequest, PutRequest, DeleteRequest

### files - File Operations for Pipes

File operations optimized for pipeline workflows:

```kukicha
import "stdlib/files"

# Read and process file
content := "input.txt"
    |> files.Read()
    |> string.Split("\n")
    |> slice.Filter(line -> not string.IsEmpty(line))
    |> string.Join("\n")
    onerr panic "read failed"

# Write data to file (auto-serializes to JSON)
repos |> files.Write("repos.json") onerr panic "write failed"

# List and filter files
logs := files.List("/var/log")
    |> slice.Filter(f -> string.HasSuffix(f.Name, ".log"))
    |> slice.Map(f -> f.Path)
```

**Functions:** Read, ReadBytes, Write, WriteString, Append, Exists, IsDir, IsFile, List, ListRecursive, Delete, DeleteAll, Copy, Move, MkDir, MkDirAll, TempFile, TempDir, Size, ModTime, Basename, Dirname, Extension, Join, Abs

### parse - Data Format Parsing

Parse common data formats (JSON, CSV, YAML) into Kukicha types:

```kukicha
import "stdlib/parse"
import "stdlib/files"

# Parse JSON from string or file
config := "config.json"
    |> files.Read()
    |> parse.Json() as Config
    onerr defaultConfig()

# Parse CSV data
records := csvData
    |> parse.Csv()
    onerr empty list of list of string

# Parse CSV with headers (returns maps)
users := csvData
    |> parse.CsvWithHeader()
    |> slice.Map(row -> User{
        Name: row["name"],
        Age: parseInt(row["age"]),
    })

# Parse YAML configuration
settings := yamlStr
    |> parse.Yaml() as Settings
    onerr defaultSettings()

# Real-world pipeline: fetch and parse
repos := fetch.Get("https://api.github.com/users/golang/repos")
    |> fetch.CheckStatus()
    |> fetch.Text()
    |> parse.Json() as list of Repo
    |> slice.Filter(r -> r.Stars > 100)
```

**Functions:** Json, Csv, CsvWithHeader, Yaml

## Special Transpilation (iter package only)

The `iter` package uses special transpilation to generate generic Go code without requiring generic syntax in Kukicha:

**Kukicha source:**
```kukicha
func Filter(seq iter.Seq, keep func(any) bool) iter.Seq
    return func(yield func(any) bool) bool
        for item in seq
            if keep(item)
                if !yield(item)
                    return false
        return true
```

**Generated Go:**
```go
func Filter[T any](seq iter.Seq[T], keep func(T) bool) iter.Seq[T] {
    return func(yield func(T) bool) bool {
        for item := range seq {
            if keep(item) {
                if !yield(item) {
                    return false
                }
            }
        }
        return true
    }
}
```

This keeps Kukicha simple while enabling type-safe generic iteration.

## What's NOT Included

These packages were considered but are better used directly from Go:

| Package | Use Instead |
|---------|-------------|
| bytes | `import "bytes"` + `onerr` |
| io | `import "io"` + `onerr` |
| time | `import "time"` + `onerr` |
| context | `import "context"` |
| json | `import "encoding/json"` + `onerr` |
| http | `import "net/http"` + `onerr` |

See [kukicha-design-philosophy.md](../docs/kukicha-design-philosophy.md) for rationale.

## Package Structure

```
stdlib/
├── fetch/         # HTTP client for pipes
│   ├── fetch.kuki
│   └── fetch_test.kuki
├── files/         # File operations for pipes
│   ├── files.kuki
│   └── files_test.kuki
├── iter/          # Iterator operations (special transpilation)
│   ├── iter.kuki
│   └── iter_test.kuki
├── parse/         # Data format parsing
│   └── parse.kuki
├── slice/         # Slice operations
│   └── slice.kuki
├── string/        # String utilities
│   └── string.kuki
└── README.md
```
