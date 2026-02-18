# Build a Concurrent URL Health Checker

**Level:** Intermediate/Advanced
**Prerequisite:** [Link Shortener Tutorial](web-app-tutorial.md)

In the previous tutorials, everything happened one step at a time: fetch a URL, parse the JSON, display the result, then move on. That works fine when you have a handful of things to do. But what if you need to check whether 100 websites are online? Doing them one at a time means waiting for each response before starting the next — most of your program's time is spent idle, waiting on the network.

In this tutorial, you'll build a **URL health checker** that checks many websites at the same time. Along the way, you'll learn:

1. **Interfaces** — defining shared behavior across different types
2. **Type assertions (`as`)** — checking what concrete type you're working with
3. **Goroutines (`go`)** — running functions in the background
4. **Channels (`send`, `receive`, `close`)** — passing data between goroutines
5. **The Fan-out pattern** — limiting how many things run at once
6. **Error wrapping** — adding context when things go wrong

---

## Step 1: The Sequential Checker

Let's start with a version that checks URLs one at a time. This gives us working code to improve later.

Create `health.kuki`:

```kukicha
import "fmt"
import "time"
import "stdlib/fetch"

type Result
    url string
    status string
    latency time.Duration

function check(url string) Result
    start := time.Now()

    resp := fetch.Get(url) onerr
        return Result{url: url, status: "DOWN ({error})", latency: time.Since(start)}

    resp = resp |> fetch.CheckStatus() onerr
        return Result{url: url, status: "ERROR ({resp.StatusCode})", latency: time.Since(start)}

    return Result{url: url, status: "UP", latency: time.Since(start)}

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
        print("[{result.status}] {result.url} ({result.latency})")
```

Walk through the `check` function:

1. `time.Now()` records when we started, so we can measure how long the check takes.
2. `fetch.Get(url)` makes an HTTP request. If the network call fails entirely (DNS error, timeout, etc.), the `onerr` block returns a `Result` with status `"DOWN"`.
3. `fetch.CheckStatus()` verifies the server returned a success code (200). If the server responded but with an error (like 500), we return `"ERROR"`.
4. If both steps succeed, the site is `"UP"`.

**Try it:**
```bash
kukicha run health.kuki
```

This works, but there's a problem. Each `fetch.Get` waits for a response before the next one starts. If you have 100 URLs and each takes 1 second, you're waiting 100 seconds total — even though your computer could easily handle all those requests at the same time.

---

## Step 2: Interfaces

Before we speed things up, let's make the design flexible. Right now we can only check HTTP URLs. But what if you later want to check whether a database is reachable, or whether a server responds to ping? All of these are "health checks" — they just check different things.

An **interface** defines a set of methods that a type must have, without saying how those methods work. It's like a job description: "must be able to `Check()` and return a `Result`."

```kukicha
interface Checker
    Check() Result
```

This says: any type that has a `Check()` method returning a `Result` counts as a `Checker`. Now let's make our HTTP check satisfy that interface:

```kukicha
type HTTPChecker
    url string

function Check on c HTTPChecker() Result
    return check(c.url)
```

`HTTPChecker` has a `Check()` method that returns a `Result`, so it automatically satisfies the `Checker` interface. No explicit "implements" declaration needed — if the methods match, the type fits.

Why bother? Because now you can write functions that work with *any* kind of checker:

```kukicha
function runCheck(c Checker)
    result := c.Check()
    print("[{result.status}] {result.url}")
```

This function doesn't care whether `c` is an `HTTPChecker`, a `PingChecker`, or something you haven't written yet. It just calls `Check()` and prints the result.

### The `as` Keyword (Type Assertions)

Sometimes you have a generic `Checker` and you need to know what specific type it is — maybe to access a field that only `HTTPChecker` has. The `as` keyword lets you check:

```kukicha
function identify(c Checker)
    http, ok := c as HTTPChecker
    if ok
        print("This is an HTTP check for {http.url}")
    else
        print("This is some other type of check")
```

`c as HTTPChecker` returns two values: the converted value and a boolean indicating whether the conversion worked. This is safe — if `c` isn't actually an `HTTPChecker`, `ok` is `false` and nothing breaks.

---

## Step 3: Goroutines

Now let's make things fast. In bash, you can run a command in the background with `&`:

```bash
curl https://google.com &
curl https://github.com &
wait
```

Kukicha has something similar but much more powerful: **goroutines**. Put `go` before a function call, and it runs in the background:

```kukicha
function main()
    urls := list of string{"https://google.com", "https://github.com"}

    for url in urls
        go check(url)

    time.Sleep(2 * time.Second)
```

Each `go check(url)` starts the `check` function running independently — it doesn't wait for it to finish before moving to the next iteration. All the checks run at the same time.

But there's a problem with this code. The `check` function returns a `Result`, but when you call it with `go`, the return value is discarded — there's nowhere for it to go. And the `time.Sleep` at the end is a hack; we're just guessing how long to wait. We need a proper way for the background work to send results back.

---

## Step 4: Channels

A **channel** is a pipe that connects goroutines. One goroutine puts data in, another takes data out. If you've used Unix pipes (`cmd1 | cmd2`), the idea is similar — except channels carry typed values, not text.

```kukicha
function main()
    urls := list of string{"https://google.com", "https://github.com"}

    results := make(channel of Result)

    for url in urls
        u := url
        go
            res := check(u)
            send res to results

    for i from 0 to len(urls)
        result := receive from results
        print("[{result.status}] {result.url}")
```

There's a lot going on here. Let's take it apart:

**Creating the channel:**
```kukicha
results := make(channel of Result)
```
This creates an **unbuffered channel** — a pipe that can carry `Result` values. "Unbuffered" means a `send` will block (wait) until another goroutine is ready to `receive`. This keeps the sender and receiver in sync.

**Launching goroutines:**
```kukicha
for url in urls
    u := url
    go
        res := check(u)
        send res to results
```

Two things to notice here:

First, `u := url` creates a local copy of the URL. This is important. Without it, all the goroutines would share the same `url` variable, and by the time they actually run, the loop might have already moved on to the next URL. Each goroutine would end up checking whatever `url` happened to be at that moment — probably the last one in the list. Copying it into `u` gives each goroutine its own value.

Second, `go` followed by an indented block starts an **anonymous goroutine** — like an inline function that runs in the background. Inside it, we call `check(u)` and then `send` the result into the channel.

**Collecting results:**
```kukicha
for i from 0 to len(urls)
    result := receive from results
    print("[{result.status}] {result.url}")
```

`receive from results` pulls one value out of the channel. If no value is available yet, it **blocks** — the program pauses at that line until a goroutine sends something. We loop exactly `len(urls)` times because we know that's how many results to expect.

No more `time.Sleep` hack. The `receive` calls naturally wait until all the work is done.

---

## Step 5: The Fan-out Pattern

The code in Step 4 launches one goroutine per URL. For 10 URLs, that's fine. For 10,000 URLs, you'd be opening 10,000 simultaneous network connections — which might overwhelm your system or get you rate-limited by the servers you're checking.

The **fan-out pattern** solves this by using a fixed number of **workers**. You put all the jobs into one channel, and the workers pull from it. Each worker handles one job at a time, picks up the next one when it's done.

Think of it like a restaurant kitchen: instead of hiring 10,000 cooks for 10,000 orders, you have 3 cooks who each grab the next ticket when they finish one.

```kukicha
function worker(id int, jobs channel of string, results channel of Result)
    for
        url := receive from jobs onerr break

        print("Worker {id} checking {url}")
        send check(url) to results
```

Walk through the worker:

1. `for` with no condition — an infinite loop. The worker keeps running until something stops it.
2. `receive from jobs` — pulls the next URL from the jobs channel. If there's nothing to do yet, it blocks and waits.
3. `onerr break` — this is how workers know when to stop. When the main function **closes** the jobs channel (meaning "no more work is coming"), the next `receive` triggers the `onerr` handler, and `break` exits the loop.
4. `send check(url) to results` — does the actual health check and pushes the result into the results channel.

Now the main function that orchestrates everything:

```kukicha
function main()
    numWorkers := 3
    urls := list of string{"https://google.com", "https://github.com", "https://go.dev"}

    jobs := make(channel of string, len(urls))
    results := make(channel of Result, len(urls))

    # Start workers
    for i from 1 through numWorkers
        go worker(i, jobs, results)

    # Send all URLs into the jobs channel
    for url in urls
        send url to jobs

    # Close the channel — tells workers "no more jobs are coming"
    close(jobs)

    # Collect results
    for i from 0 to len(urls)
        res := receive from results
        print("Done: {res.url}")
```

A few new details:

**Buffered channels:**
```kukicha
jobs := make(channel of string, len(urls))
results := make(channel of Result, len(urls))
```
The second argument to `make` is the **buffer size**. An unbuffered channel (Step 4) blocks on `send` until someone is ready to `receive`. A buffered channel can hold that many values before blocking. Here we make the buffer big enough to hold all the URLs, so the main function can send all the jobs without waiting for workers to pick them up.

**Closing the channel:**
```kukicha
close(jobs)
```
This signals "no more values will be sent on this channel." Workers that are waiting on `receive from jobs` will get the signal and their `onerr break` will fire. Without `close`, the workers would wait forever for more jobs that never come, and your program would hang.

The flow looks like this:
1. Start 3 workers (they immediately block on `receive from jobs` — nothing to do yet).
2. Send all URLs into the jobs channel.
3. Close the jobs channel.
4. Workers wake up, each grabs a URL, checks it, sends the result, grabs the next URL, and so on until the channel is empty and closed.
5. Main collects all results.

With 3 workers and 100 URLs, at most 3 checks happen simultaneously — controlled and predictable.

---

## Step 6: Logging Results

For a real health checker, you'd want to log results to a file so you can review them later. Let's add that:

```kukicha
import "stdlib/files"
import "stdlib/datetime"
import "stdlib/errors"

function logResult(res Result)
    now := datetime.Now() |> datetime.Format(datetime.RFC3339)
    line := "{now} | [{res.status}] {res.url} | {res.latency}\n"

    files.AppendString("health.log", line) onerr
        print(errors.Wrap(error, "log write failed"))
```

`errors.Wrap(error, "log write failed")` adds context to the error — you'll see something like `"log write failed: permission denied"` instead of just `"permission denied"`. This is the same pattern from [Tutorial 2](data-scripting-tutorial.md).

To use this in the fan-out version, add a `logResult` call after each check completes — either inside the worker, or in the main function's result collection loop.

---

## Summary

Here's what you learned in this tutorial:

| Concept | Syntax | What It Does |
|---------|--------|--------------|
| **Interface** | `interface Checker` | Defines methods a type must have |
| **Type assertion** | `c as HTTPChecker` | Checks if a value is a specific type |
| **Goroutine** | `go check(url)` | Runs a function in the background |
| **Channel** | `make(channel of Result)` | A pipe for passing data between goroutines |
| **Send** | `send res to results` | Puts a value into a channel |
| **Receive** | `receive from results` | Takes a value out (blocks until one arrives) |
| **Close** | `close(jobs)` | Signals no more values will be sent |
| **Buffered channel** | `make(channel of T, 10)` | Channel that can hold values before blocking |
| **Error wrapping** | `errors.Wrap(error, "msg")` | Adds context to an error |

### What's Next?

You've built something that runs real work in parallel — one of the things that makes compiled languages worth the switch from scripting. Next, learn how to build production-grade applications:

| # | Tutorial | What You'll Learn |
|---|----------|-------------------|
| 1 | **[Beginner Tutorial](beginner-tutorial.md)** | Variables, functions, strings, decisions, lists, loops, pipes |
| 2 | **[Data & AI Scripting](data-scripting-tutorial.md)** | Maps (Key-Value), parsing CSVs, shell commands, AI scripting |
| 3 | **[CLI Explorer](cli-explorer-tutorial.md)** | Custom types, methods, API data, arrow lambdas, error handling |
| 4 | **[Link Shortener](web-app-tutorial.md)** | HTTP servers, JSON, REST APIs, redirects |
| 5 | **You are here** | Concurrency, goroutines, channels, interfaces |
| 6 | **[Production Patterns](production-patterns-tutorial.md)** | Databases, advanced patterns |
