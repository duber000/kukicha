# Go Code Generation: Known Issues

Audit of the generated `.go` files in `stdlib/`. These are patterns the Kukicha compiler produces that a Go programmer would flag.

---

## Bugs

### 1. `shell.buildExecCmd` — defer cancels context before command runs

**File:** `stdlib/shell/shell.go` (generated from `shell.kuki:103-109`)

```go
func buildExecCmd(cmd Command) *exec.Cmd {
    if cmd.timeout > 0 {
        h := ctxpkg.WithTimeout(ctxpkg.Background(), int64(cmd.timeout))
        defer ctxpkg.Cancel(h)  // fires when buildExecCmd returns, not after Run()
        return exec.CommandContext(ctxpkg.Value(h), cmd.name, cmd.args...)
    }
    return exec.Command(cmd.name, cmd.args...)
}
```

The `defer` fires when `buildExecCmd` returns — **before `execCmd.Run()` is called** in `Execute`. The timeout context is dead on arrival. The cancel should happen after the command finishes.

### 2. `fetch.doOnce` — response body leak with `maxBodySize`

**File:** `stdlib/fetch/fetch.go` (generated from `fetch.kuki:138-139`)

```go
resp.Body = io.NopCloser(io.LimitReader(resp.Body, req.maxBodySize))
```

The original `resp.Body` is replaced with a `NopCloser`-wrapped `LimitReader`. When the caller closes the new body, the original body's `Close()` is never called — the underlying TCP connection leaks. Should wrap so that `Close` still delegates to the original body.

### 3. `concurrent.Parallel` / `concurrent.Map` — panic hangs forever

**File:** `stdlib/concurrent/concurrent.go`

```go
go func() {
    t()
    wg.Done()  // never reached if t() panics
}()
```

If `t()` panics, `wg.Done()` is never called and `wg.Wait()` blocks forever. The standard fix is `defer wg.Done()`. Affects `Parallel`, `ParallelWithLimit`, `Map`, and `MapWithLimit`.

---

## Code Smells

### 4. `errors.New(fmt.Sprintf(...))` instead of `fmt.Errorf`

**Files:** `fetch.go`, `shell.go`, `pg.go`, `slice.go`, and others

```go
// Generated:
return nil, errors.New(fmt.Sprintf("request failed: status %v", resp.StatusCode))

// Idiomatic Go:
return nil, fmt.Errorf("request failed: status %d", resp.StatusCode)
```

This is a well-known Go anti-pattern. `fmt.Errorf` does the same thing in one call and supports `%w` for error wrapping.

### 5. `*new(T)` for generic zero values

**Files:** `slice.go`, `pg.go`, `fetch.go`, `iterator.go`

```go
// Generated:
return *new(T), errors.New("slice is empty")

// Idiomatic Go:
var zero T
return zero, errors.New("slice is empty")
```

`*new(T)` heap-allocates a `T` then immediately dereferences it. A `var zero T` declaration avoids the allocation entirely.

### 6. `pg` types store concrete values as `any`

**File:** `stdlib/pg/pg.go`

```go
type Row struct  { scanFn any }
type Rows struct { rows any }
type Tx struct   { tx any }
```

These are then bare-asserted everywhere (e.g. `r.scanFn.(pgx.Row)`). If the assertion fails, it panics at runtime. These should store the concrete types (`pgx.Row`, `pgx.Rows`, `pgx.Tx`) directly.

### 7. `fetch.createHTTPRequest` — bare type assertion

**File:** `stdlib/fetch/fetch.go`

```go
bodyBytes := bodyData.([]byte)  // panics if bodyData isn't []byte
```

Should use the comma-ok form or ensure the caller guarantees the type.

### 8. `slice.Concat` — item-by-item append

**File:** `stdlib/slice/slice.go`

```go
// Generated (capacity is pre-allocated, but appends one at a time):
for _, slice := range slices {
    for _, item := range slice {
        result = append(result, item)
    }
}

// Idiomatic Go:
for _, s := range slices {
    result = append(result, s...)
}
```

### 9. `slice.Unique` — `map[K]bool` instead of `map[K]struct{}`

**File:** `stdlib/slice/slice.go`

```go
// Generated:
seen := make(map[K]bool)

// Idiomatic Go (zero-cost values):
seen := make(map[K]struct{})
```

### 10. Magic numbers for retry strategy

**Files:** `stdlib/retry/retry.go`, `stdlib/fetch/fetch.go`, `stdlib/pg/pg.go`

```go
cfg := retry.Config{MaxAttempts: 3, InitialDelay: 1000, Strategy: 1}  // what is 1?
```

`Strategy: 0` means linear, `Strategy: 1` means exponential. Should be named constants.

### 11. `retry.calculateDelay` — integer overflow

**File:** `stdlib/retry/retry.go`

```go
multiplier := 1
for range attempt {
    multiplier = (multiplier * 2)
}
```

At attempt 63+, `multiplier` overflows `int`. Should cap the maximum delay.

---

## Design Concerns

### 12. `files.Watch` — no cancellation mechanism

**File:** `stdlib/files/files.go`

`Watch` runs an infinite `for {}` loop with `time.Sleep(500ms)`. There is no context, done channel, or any way for the caller to stop it. The goroutine is trapped forever.

### 13. `files.Copy` — doesn't preserve file permissions

**File:** `stdlib/files/files.go`

```go
destFile, err_15 := os.Create(dst)  // always mode 0666 (minus umask)
```

The source file's permissions are lost. Should `os.Stat` the source and apply the same mode.

### 14. `fetch.DownloadTo` — unnecessary string conversion doubles memory

**File:** `stdlib/fetch/fetch.go`

```go
bodyBytes, err_16 := io.ReadAll(resp.Body)
return sandbox.WriteString(box, string(bodyBytes), path)  // copies all bytes again
```

For large downloads this doubles peak memory. Should write bytes directly.

### 15. `fetch.NewSession` — swallowed error

**File:** `stdlib/fetch/fetch.go`

```go
jar, _ := cookiejar.New(nil)
```

While `cookiejar.New(nil)` never errors in practice, discarding errors is a Go review red flag.

---

## What Looks Good

- **Generics** in `slice/`, `iterator/`, `sort/` are clean and idiomatic
- **`iter.Seq` iterators** correctly handle early termination via `yield` return values
- **Error wrapping** with `%w` in `pg/` is mostly done right
- **Builder pattern** using value receiver copies (`fetch/`, `shell/`, `pg/`) is smart — mutation is safe without pointer aliasing
- **Security helpers** (`SafeRedirect`, `SafeGet`, `SafeHTML`, `SetSecureHeaders`) are well-designed
