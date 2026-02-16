# Concurrent URL Health Checker

**Level:** Advanced
**Time:** 40 minutes
**Prerequisite:** [Link Shortener Tutorial](web-app-tutorial.md)

In this tutorial, we'll build a high-performance **URL Health Checker**. You'll learn how to monitor thousands of websites simultaneously using Kukicha's powerful concurrency primitives.

## What You'll Learn

- How to use **Interfaces** to build flexible, swappable components
- How to use the **`as` keyword** for type assertions
- How to launch **Goroutines** with the `go` keyword
- How to coordinate work using **Channels** (`send`, `receive`, `close`)
- How to implement the **Fan-out** pattern for parallel processing
- How to wrap errors with context using **`stdlib/errors`**

---

## Step 1: The Sequential Checker

Let's start by building a simple version that checks URLs one by one.

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

**The Problem:** If you have 100 URLs and each takes 1 second, this will take 100 seconds. We can do better!

---

## Step 2: Interfaces & Type Assertions

Before we go fast, let's make our code flexible. What if we want to check more than just HTTP? Maybe we want to check Ping or a Database connection.

We'll define an **`interface`**:

```kukicha
interface Checker
    Check() Result
```

And a specific **HTTPChecker**:

```kukicha
type HTTPChecker
    url string

func Check on c HTTPChecker() Result
    return check(c.url) # Uses the function we wrote earlier
```

### The `as` Keyword (Type Assertions)

Sometimes you have a generic `Checker` and you need to know if it's specifically an `HTTPChecker` to access its unique fields. This is where **`as`** comes in:

```kukicha
function identify(c Checker)
    # Type assertion: "is this a reference to an HTTPChecker?"
    http, ok := c as reference HTTPChecker
    if ok
        print("This is an HTTP check for {http.url}")
    else
        print("This is some other type of check")
```

In Kukicha, `c as T` returns two values: the converted value and a boolean `ok` indicating success.

---

## Step 3: Goroutines (`go`)

Kukicha (like Go) makes concurrency easy with **goroutines**. To run a function in the background, just put **`go`** before it.

```kukicha
function main()
    urls := list of string{"https://google.com", "https://github.com"}

    for url in urls
        # Start a goroutine for each check
        go check(url)
    
    # Wait a bit so the program doesn't exit immediately
    time.Sleep(2 * time.Second)
```

**Wait!** The program finished before the checks could print anything. We need a way to communicate back to the main thread.

---

## Step 4: Channels (`send`, `receive`)

**Channels** are the pipes that connect goroutines. One goroutine **sends** data, and another **receives** it.

```kukicha
function main()
    urls := list of string{"https://google.com", "https://github.com"}
    
    # Create a channel that carries 'Result' types
    results := make channel of Result
    
    for url in urls
        u := url # Shadow the variable for the goroutine
        go
            res := check(u)
            # Send the result into the channel
            send res to results
    
    # Receive the results
    for i from 0 to len(urls)
        # Program blocks here until a message arrives
        result := receive from results
        print("[{result.status}] {result.url}")
```

### Key Concepts:
1. **`make channel of T`**: Creates a new channel.
2. **`send channel, value`**: Pushes a value into the pipe.
3. **`receive from channel`**: Pulls a value out. This blocks the current goroutine until something is available.

---

## Step 5: The Fan-out Pattern

Launching 10,000 goroutines for 10,000 URLs is fine for local tests, but in production, you might want to limit yourself to eg. 10 workers at a time. This is called **Fan-out**.

```kukicha
func worker(id int, jobs channel of string, results channel of Result)
    for
        # Receive a URL from the jobs channel
        url := receive from jobs onerr break # Exit loop if channel is closed
        
        print("Worker {id} checking {url}")
        send check(url) to results

function main()
    numWorkers := 3
    urls := list of string{"https://google.com", "https://github.com", "https://go.dev"}

    jobs := make(channel of string, len(urls))
    results := make(channel of Result, len(urls))

    # Start workers
    for i from 1 through numWorkers
        go worker(i, jobs, results)

    # Fill the 'jobs' pipe
    for url in urls
        send url to jobs
    
    # Close the jobs channel so workers know to stop
    close(jobs)

    # Collect results
    for i from 0 to len(urls)
        res := receive from results
        print("Done: {res.url}")
```

---

## Step 6: Logging and Checksums (Production Patch)

Finally, let's log these results to a file with timestamps.

```kukicha
import "stdlib/files"
import "stdlib/datetime"
import "stdlib/errors"

function logResult(res Result)
    now := datetime.Now() |> datetime.Format(datetime.RFC3339)
    line := "{now} | [{res.status}] {res.url} | {res.latency}\n"

    # Append to log file — wrap the error with context if it fails
    files.AppendString("health.log", line) onerr
        print(errors.Wrap(error, "log write failed"))
```

---

## Summary

Congratulations! You've built a concurrent system in Kukicha.

| Feature | Description |
|---------|-------------|
| **`interface`** | Defines behavior without implementation |
| **`as`** | Checks and converts types safely |
| **`go`** | Starts a light-weight background thread (goroutine) |
| **`channel`** | Safe communication between goroutines |
| **`send`/`receive`** | Moving data through channels |
| **`stdlib/errors`** | Wrap errors with context (`errors.Wrap`) |

### Next Steps

Now that you've mastered concurrency, move on to the [Production Patterns Tutorial](production-patterns-tutorial.md) to learn about databases and advanced error handling!
