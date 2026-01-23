# Kukicha Standard Library Roadmap

**Version:** 2.0.0
**Status:** Stable (Philosophy Shift)
**Updated:** 2026-01-22

---

## Design Philosophy: "It's Just Go"

Following CoffeeScript's successful approach with JavaScript, Kukicha adopts the principle: **"It's just Go."**

Kukicha provides syntactic sugar for Go, not a parallel standard library. You can use any Go package directly, and `onerr` handles error tuples at the call site.

**Key Principles:**
- Go stdlib is first-class in Kukicha
- Use `onerr` for error handling, not wrapper functions
- Kukicha stdlib only for genuinely new capabilities Go lacks
- No maintenance burden from tracking Go stdlib evolution

See [kukicha-design-philosophy.md](kukicha-design-philosophy.md) for full rationale.

---

## Implementation Status

### ✅ Completed Packages (Value-Add)

These packages provide functionality Go lacks or make awkward:

| Package | Functions | Purpose |
|---------|-----------|---------|
| **iter** | 13 | Functional iteration (Filter, Map, Reduce) |
| **slice** | 12 | Slice operations (First, Last, Drop, Unique) |
| **string** | 28 | String utilities (thin wrappers, low maintenance) |

### ❌ Deprecated Plans

The following were previously planned but are **no longer needed** under the CoffeeScript model:

| Package | Reason | Use Instead |
|---------|--------|-------------|
| bytes | Go stdlib works directly | `import "bytes"` + `onerr` |
| io | Go stdlib works directly | `import "io"` + `onerr` |
| time | Go stdlib works directly | `import "time"` + `onerr` |
| context | Go stdlib works directly | `import "context"` |
| url | Go stdlib works directly | `import "net/url"` + `onerr` |
| json | Go stdlib works directly | `import "encoding/json"` + `onerr` |
| http | Go stdlib works directly | `import "net/http"` + `onerr` |

---

## Core Libraries

### Iter Package ✅ COMPLETED

Functional iteration operations using Go 1.23+ `iter.Seq`. Go's iterator protocol is low-level; we provide higher-level operations.

```kukicha
import "stdlib/iter"
import "slices"

# Lazy filtering and transformation
result := slices.Values(numbers)
    |> iter.Filter(func(n int) bool { return n > 0 })
    |> iter.Map(func(n int) int { return n * 2 })
    |> iter.Collect()

# Available functions:
# - Filter, Map, FlatMap
# - Take, Skip
# - Enumerate, Zip
# - Chunk
# - Reduce, Collect
# - Any, All, Find
```

### Slice Package ✅ COMPLETED

Extended slice operations Go lacks:

```kukicha
import "stdlib/slice"

# Take/drop operations
firstThree := slice.First(items, 3)
lastTwo := slice.Last(items, 2)
tail := slice.Drop(items, 3)
head := slice.DropLast(items, 1)

# Pipeline-friendly
result := items
    |> slice.Drop(2)
    |> slice.First(10)
    |> process()

# Additional operations
reversed := slice.Reverse(items)
unique := slice.Unique(items)
chunked := slice.Chunk(items, 5)
```

### String Package ✅ COMPLETED

String utilities (thin wrappers with minimal maintenance burden):

```kukicha
import "stdlib/string"

# Case conversion
upper := string.ToUpper("hello")
lower := string.ToLower("WORLD")

# Trimming
trimmed := string.TrimSpace("  hello  ")
withoutPrefix := string.TrimPrefix(url, "https://")

# Splitting and joining
parts := string.Split("a,b,c", ",")
joined := string.Join(parts, "|")

# Searching
if string.Contains(text, "error")
    handleError()
```

---

## Using Go Standard Library Directly

Use Go packages directly with pure Kukicha syntax:

### JSON

```kukicha
import "encoding/json"

type User struct
    Name  string
    Email string

# Marshal with error handling
func SaveUser(user User) error
    data := json.Marshal(user) onerr return error
    return os.WriteFile("user.json", data, 0644)

# Unmarshal - using 'reference of' for address-of
func LoadUser(path string) (User, error)
    data := os.ReadFile(path) onerr return User{}, error
    user := User{}
    json.Unmarshal(data, reference of user) onerr return User{}, error
    return user, empty

# Convenience pattern - panic for scripts
func MustLoadUser(path string) User
    data := os.ReadFile(path) onerr panic "cannot read file"
    user := User{}
    json.Unmarshal(data, reference of user) onerr panic "invalid json"
    return user
```

### HTTP

```kukicha
import "net/http"
import "encoding/json"
import "io"

# Simple GET
func FetchData(url string) (list of byte, error)
    resp := http.Get(url) onerr return empty, error
    defer resp.Body.Close()
    return io.ReadAll(resp.Body)

# GET with JSON decoding
func FetchUsers(url string) (list of User, error)
    resp := http.Get(url) onerr return empty, error
    defer resp.Body.Close()

    users := list of User{}
    decoder := json.NewDecoder(resp.Body)
    decoder.Decode(reference of users) onerr return empty, error
    return users, empty

# POST with JSON body
func CreateUser(url string, user User) error
    data := json.Marshal(user) onerr return error
    resp := http.Post(url, "application/json", bytes.NewReader(data)) onerr return error
    defer resp.Body.Close()

    if resp.StatusCode >= 400
        return fmt.Errorf("request failed: {resp.Status}")
    return empty
```

### HTTP Server

```kukicha
import "net/http"
import "encoding/json"

func main()
    http.HandleFunc("/users", handleUsers)
    http.HandleFunc("/users/", handleUser)  # Go 1.22+ path patterns

    print "Server starting on :8080"
    http.ListenAndServe(":8080", empty) onerr panic "server failed"

func handleUsers(w http.ResponseWriter, r reference http.Request)
    if r.Method equals "GET"
        users := getAllUsers()
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(users)
    else if r.Method equals "POST"
        user := User{}
        json.NewDecoder(r.Body).Decode(reference of user) onerr
            http.Error(w, "Invalid JSON", 400)
            return
        saveUser(user)
        w.WriteHeader(201)
```

### File I/O

```kukicha
import "os"
import "bufio"

# Read entire file
func ReadConfig(path string) (list of byte, error)
    return os.ReadFile(path)

# Write file
func WriteOutput(path string, data list of byte) error
    return os.WriteFile(path, data, 0644)

# Read lines
func ReadLines(path string) (list of string, error)
    file := os.Open(path) onerr return empty, error
    defer file.Close()

    lines := list of string{}
    scanner := bufio.NewScanner(file)
    for scanner.Scan()
        lines = append(lines, scanner.Text())
    return lines, scanner.Err()
```

### Time and Context

```kukicha
import "time"
import "context"

# Timeouts
func FetchWithTimeout(url string) (list of byte, error)
    ctx, cancel := context.WithTimeout(context.Background(), 30 * time.Second)
    defer cancel()

    req := http.NewRequestWithContext(ctx, "GET", url, empty) onerr return empty, error
    resp := http.DefaultClient.Do(req) onerr return empty, error
    defer resp.Body.Close()
    return io.ReadAll(resp.Body)

# Periodic tasks
func StartWorker(interval time.Duration)
    ticker := time.NewTicker(interval)
    defer ticker.Stop()

    for t in ticker.C
        print "Tick at {t}"
        doWork()
```

---

## Future Directions

### Potential Value-Add Packages

If compelling use cases emerge, we may add:

| Package | Potential Purpose |
|---------|------------------|
| **maps** | Map operations Go lacks (Filter, Keys, Values, Merge) |
| **result** | Optional Result type for explicit error handling |

### Not Planned

These remain accessible via direct Go imports:

- Cloud SDKs (use AWS, GCP, Azure Go SDKs directly)
- Database drivers (use standard Go drivers)
- Testing (use Go's `testing` package)
- AI/LLM clients (use official Go SDKs)

---

## Rationale: Why Not Wrappers?

The previous roadmap planned ~187 wrapper functions. This was abandoned because:

1. **Maintenance burden**: Go stdlib evolves; wrappers become stale
2. **Incomplete coverage**: Always missing some function
3. **Documentation overhead**: Two sets of docs to maintain
4. **Blocking issues**: Tuple returns required compiler changes
5. **No real benefit**: `onerr` handles errors elegantly already

CoffeeScript thrived by being "just JavaScript" - no parallel ecosystem. Kukicha follows the same path.

---

## Contributing

The stdlib is largely complete. Future contributions should focus on:

1. **Improving existing packages** (iter, slice, string)
2. **Documentation and examples** for using Go stdlib in Kukicha
3. **Compiler improvements** (onerr patterns, syntax sugar)

For significant new stdlib packages, open an issue first to discuss whether it provides genuine value Go lacks.

---

**Last Updated:** 2026-01-22
**Philosophy Document:** [kukicha-design-philosophy.md](kukicha-design-philosophy.md)
