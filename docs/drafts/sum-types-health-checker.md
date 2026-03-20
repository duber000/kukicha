# Draft: Sum Types in the Health Checker Tutorial

> This is a draft showing how sum types would change the concurrent URL health checker tutorial.
> It's written as replacement/additional sections, not a full tutorial rewrite.

---

## The Problem With String Status

The current health checker uses a string to represent what happened:

```kukicha
type Result
    url string
    status string          # "UP", "DOWN (...)", "ERROR (500)"
    latency time.Duration
```

This works, but it has a subtle problem. When you write code that *uses* a Result, you're pattern-matching on strings:

```kukicha
# What if someone adds a new status like "TIMEOUT" later?
# This code silently ignores it — no compiler warning.
if result.status |> string.HasPrefix("DOWN")
    alertOncall(result.url)
else if result.status |> string.HasPrefix("ERROR")
    logHttpError(result.url)
# Forgot to handle "TIMEOUT" — nobody finds out until production
```

The compiler can't help you because it doesn't know what values `status` can hold. It's just a string.

---

## Sum Types: The Compiler Knows Your Variants

A **sum type** declares all the possible variants of a value in one place. The compiler then enforces that every `switch` handles all of them.

```kukicha
type HealthStatus
    = Up
    | Down(reason string)
    | HttpError(code int)

type Result
    url string
    status HealthStatus
    latency time.Duration
```

Read this as: "A `HealthStatus` is **either** `Up`, **or** `Down` with a reason, **or** an `HttpError` with a status code — and nothing else."

Each variant can carry different data:
- `Up` carries nothing — the site is healthy, that's all we need to know
- `Down` carries a `reason` string — why the connection failed (DNS error, timeout, etc.)
- `HttpError` carries a `code` int — the HTTP status code the server returned (500, 403, etc.)

### The `|` separator

The `|` between variants means "or" — it separates the alternatives. This is different from `|>` (the pipe operator that chains function calls). You'll only ever see `|` inside a `type ... =` declaration, never in expressions:

```kukicha
# |  means "or" — separates variants in a type declaration
type Light = Red | Yellow | Green

# |> means "pipe into" — chains function calls in expressions
data |> parse() |> transform()
```

---

## Rewriting the Check Function

Here's how `check` changes:

**Before (string status):**
```kukicha
function check(url string) Result
    start := time.Now()

    resp := fetch.Get(url) onerr
        return Result{url: url, status: "DOWN ({error})", latency: time.Since(start)}

    resp = resp |> fetch.CheckStatus() onerr
        return Result{url: url, status: "ERROR ({resp.StatusCode})", latency: time.Since(start)}

    return Result{url: url, status: "UP", latency: time.Since(start)}
```

**After (sum type status):**
```kukicha
function check(url string) Result
    start := time.Now()

    resp := fetch.Get(url) onerr
        return Result{url: url, status: Down{reason: "{error}"}, latency: time.Since(start)}

    resp = resp |> fetch.CheckStatus() onerr
        return Result{url: url, status: HttpError{code: resp.StatusCode}, latency: time.Since(start)}

    return Result{url: url, status: Up{}, latency: time.Since(start)}
```

The structure is the same — `onerr` still handles each failure mode. The difference is that each status is now a **typed variant** instead of a formatted string. The meaning is encoded in the type, not in string formatting conventions.

---

## Exhaustive Switching

Here's where it pays off. When you process results, the compiler checks that you handle every variant:

```kukicha
function display(result Result)
    result.status |> switch
        when Up
            print("[UP] {result.url} ({result.latency})")
        when Down
            print("[DOWN] {result.url} - {result.status.reason}")
        when HttpError
            print("[HTTP {result.status.code}] {result.url}")
```

Now suppose you add a `Timeout` variant six months later:

```kukicha
type HealthStatus
    = Up
    | Down(reason string)
    | HttpError(code int)
    | Timeout(after time.Duration)    # new variant
```

The compiler immediately flags every `switch` on `HealthStatus` that doesn't handle `Timeout`:

```
health.kuki:42: non-exhaustive switch on HealthStatus — missing variant: Timeout
health.kuki:78: non-exhaustive switch on HealthStatus — missing variant: Timeout
```

With the string version, adding a new status like `"TIMEOUT (...)"` compiles fine and silently falls through to `otherwise` (or gets missed entirely). With sum types, forgetting is a **compile error**, not a production bug.

---

## Replacing the Checker Interface

The tutorial currently uses an interface to support different check types:

**Before (interface + separate structs):**
```kukicha
interface Checker
    Check() Result

type HTTPChecker
    url string

function Check on c HTTPChecker() Result
    return check(c.url)

# To add a new checker, define a new struct + method anywhere.
# Nothing ties the checkers together — the compiler doesn't know what types
# implement Checker, so type-switches always need an `otherwise` branch.
```

**After (sum type — closed set of variants):**
```kukicha
type Checker
    = HTTP(url string)
    | TCP(host string, port int)
    | DNS(domain string)

function runCheck(c Checker) Result
    start := time.Now()
    c |> switch
        when HTTP
            resp := fetch.Get(c.url) onerr
                return Result{url: c.url, status: Down{reason: "{error}"}, latency: time.Since(start)}
            resp = resp |> fetch.CheckStatus() onerr
                return Result{url: c.url, status: HttpError{code: resp.StatusCode}, latency: time.Since(start)}
            return Result{url: c.url, status: Up{}, latency: time.Since(start)}
        when TCP
            addr := "{c.host}:{c.port}"
            conn := net.DialTimeout("tcp", addr, 5 * time.Second) onerr
                return Result{url: addr, status: Down{reason: "{error}"}, latency: time.Since(start)}
            conn.Close()
            return Result{url: addr, status: Up{}, latency: time.Since(start)}
        when DNS
            _ := net.LookupHost(c.domain) onerr
                return Result{url: c.domain, status: Down{reason: "{error}"}, latency: time.Since(start)}
            return Result{url: c.domain, status: Up{}, latency: time.Since(start)}
```

**When to use which:**

| Use an **interface** when... | Use a **sum type** when... |
|------------------------------|---------------------------|
| Third-party code should add new types | You know all the variants up front |
| You want open extension (plugin model) | You want exhaustiveness checking |
| Types share behavior but differ in implementation | Types share a role but carry different data |

For a health checker, the set of check types is usually known at compile time — you're not shipping a plugin API. A sum type is the better fit. Interfaces remain valuable for things like `io.Reader` where anyone should be able to implement the contract.

---

## Full Revised Step 1

Putting it all together, here's what the revised Step 1 looks like:

```kukicha
import "time"
import "stdlib/fetch"

type HealthStatus
    = Up
    | Down(reason string)
    | HttpError(code int)

type Result
    url string
    status HealthStatus
    latency time.Duration

function check(url string) Result
    start := time.Now()

    resp := fetch.Get(url) onerr
        return Result{url: url, status: Down{reason: "{error}"}, latency: time.Since(start)}

    resp = resp |> fetch.CheckStatus() onerr
        return Result{url: url, status: HttpError{code: resp.StatusCode}, latency: time.Since(start)}

    return Result{url: url, status: Up{}, latency: time.Since(start)}

function formatStatus(result Result) string
    return result.status |> switch
        when Up
            return "[UP]"
        when Down
            return "[DOWN: {result.status.reason}]"
        when HttpError
            return "[HTTP {result.status.code}]"

function main()
    urls := list of string{
        "https://google.com",
        "https://github.com",
        "https://go.dev",
        "https://invalid-url-example.test",
    }

    print("Checking {len(urls)} URLs sequentially...")

    for url in urls
        result := check(url)
        print("{formatStatus(result)} {result.url} ({result.latency})")
```

**What changed:**
- `status string` became `status HealthStatus` — three explicit variants instead of magic strings
- `formatStatus` uses an exhaustive switch — add a variant, compiler tells you to update this function
- The `check` function constructs typed variants (`Down{reason: ...}`) instead of formatting strings (`"DOWN (...)"`)
- The rest of the tutorial (goroutines, channels, fan-out) stays exactly the same — sum types only change the data modeling, not the concurrency patterns

---

## Logging With Sum Types

The Step 6 logging function also gets cleaner:

**Before:**
```kukicha
function logResult(res Result)
    now := datetime.Now() |> datetime.Format(datetime.RFC3339)
    line := "{now} | [{res.status}] {res.url} | {res.latency}\n"
    files.AppendString("health.log", line) onerr
        print(errors.Wrap(error, "log write failed"))
```

**After:**
```kukicha
function logResult(res Result)
    now := datetime.Now() |> datetime.Format(datetime.RFC3339)
    severity := res.status |> switch
        when Up
            return "INFO"
        when Down
            return "ERROR"
        when HttpError
            return "WARN"
    line := "{now} | {severity} | {formatStatus(res)} {res.url} | {res.latency}\n"
    files.AppendString("health.log", line) onerr
        print(errors.Wrap(error, "log write failed"))
```

With strings you'd have to parse the status to decide severity. With sum types, you switch on it directly — and the compiler ensures you assign a severity for every variant.
