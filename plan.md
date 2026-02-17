# Remediation Plan

## P0 - Fix symlink/path-traversal vulnerabilities in container tar helpers

**Files:** `stdlib/container/container_helper.go`

Three functions have symlink-related security issues with zero test coverage:

### `createTarFromPath` (used by `CopyTo`/`CopyToCtx`)
- `filepath.Walk()` follows symlinks silently — a symlink to `/etc/shadow` in the source tree gets archived and sent into the container.
- **Fix:** Replace `os.Stat` with `os.Lstat` at the entry point. Inside the walk callback, check `fi.Mode()&os.ModeSymlink != 0` and skip (or error). Alternatively, switch to `filepath.WalkDir` and check `d.Type()&os.ModeSymlink`.

### `extractTar` (used by `CopyFrom`/`CopyFromCtx`)
- Only handles `tar.TypeDir` and `tar.TypeReg` — symlink entries (`tar.TypeSymlink`, `tar.TypeLink`) are silently dropped. If they weren't dropped, the existing path-traversal check wouldn't catch them because it validates `header.Name` but not `header.Linkname`.
- **Fix:** Add explicit cases for `tar.TypeSymlink` and `tar.TypeLink` that either (a) reject them with an error, or (b) validate `header.Linkname` resolves within `destPath` before creating. Option (a) is simpler and safer for a DevOps tool. Also add a `default` case that returns an error for unexpected typflags (device files, etc.).

### `buildImage`
- `filepath.WalkDir()` follows symlinks in the Docker build context — same issue as `createTarFromPath`.
- **Fix:** Same approach: check `d.Type()&os.ModeSymlink` in the walk callback and skip.

### Tests
- Add `container_helper_test.go` with cases for:
  - Symlink in source tree (should be skipped/rejected)
  - Tar archive containing `tar.TypeSymlink` entry (should be rejected)
  - Tar archive with `../` path traversal (should be rejected — verify existing check)
  - Normal file/dir round-trip (sanity check)

---

## P1 - Consolidate `*Ctx` function duplication

**Files:** `stdlib/container/container_helper.go`, `stdlib/kube/kube_helper.go`

10 `*Ctx` functions double the API surface. Two distinct patterns exist:

### Pattern A — Context-only (consolidatable): 5 pairs
These pairs are identical except one uses `context.Background()` and the other uses `ctxpkg.Value(h)`:

| Module | Base | Ctx variant |
|--------|------|-------------|
| container | `Pull` | `PullCtx` |
| container | `Exec` | `ExecCtx` |
| container | `CopyFrom` | `CopyFromCtx` |
| container | `CopyTo` | `CopyToCtx` |
| kube | `PodLogs` | `PodLogsCtx` |

**Fix:** Collapse each pair into a single function that takes a `context.Context` internally. The base function becomes the single entry point and creates a `context.Background()` when no handle is provided. Since these are in `_helper.go` (raw Go, not transpiled), use a variadic `...ctxpkg.Handle` pattern:

```go
func Pull(engine Engine, ref string, handles ...ctxpkg.Handle) (string, error) {
    ctx := context.Background()
    if len(handles) > 0 {
        ctx = ctxpkg.Value(handles[0])
    }
    // ... single implementation
}
```

Delete the `PullCtx`, `ExecCtx`, `CopyFromCtx`, `CopyToCtx`, `PodLogsCtx` functions entirely. No callers exist in the codebase.

### Pattern B — Timeout-based (keep separate): 5 pairs
These have fundamentally different signatures (`timeoutSeconds int64` vs `ctxpkg.Handle`):

- `Wait`/`WaitCtx`, `Events`/`EventsCtx`
- `WaitDeploymentReady`/`WaitDeploymentReadyCtx`, `WaitPodReady`/`WaitPodReadyCtx`, `WatchPods`/`WatchPodsCtx`

**Fix:** Keep both variants but rename the timeout-based ones to communicate intent better. The `*Ctx` naming is fine for these since they genuinely have different calling conventions. No code change needed — just verify docs are clear.

---

## P2 - Add cancel/leak safety to ctx.Handle

**Files:** `stdlib/ctx/ctx.kuki`, `stdlib/ctx/ctx.go`

### Problem
`ctx.WithTimeoutMs` and `ctx.WithDeadlineUnix` return a `Handle` with a cancel function. If the user never calls `ctx.Cancel(h)`, the internal timer goroutine leaks until it fires. Go docs strongly emphasize `defer cancel()`. The `Handle` wrapper hides this footgun.

### Fix
1. Add a comment to `WithTimeoutMs` and `WithDeadlineUnix` in `ctx.kuki`:
   ```kukicha
   # IMPORTANT: Always call ctx.Cancel(handle) when done (or use defer logic)
   # to avoid resource leaks.
   ```

2. Add a `WithTimeout` convenience that takes seconds (int64) instead of milliseconds — milliseconds is an unusual default for a DevOps scripting language where most timeouts are 5s, 30s, 300s. Keep `WithTimeoutMs` for precision cases.

3. Consider whether `ctx.Cancel` on a no-cancel handle should return a `bool` instead of being a silent no-op. This is a minor API tweak — low risk since there are no callers yet.

---

## P3 - Track `_helper.go` growth; migrate what's expressible to `.kuki`

**Files:** `stdlib/container/container_helper.go` (814 lines), `stdlib/kube/kube_helper.go` (600+ lines)

### Problem
These files bypass the transpiler entirely. They're necessary for SDK calls Kukicha can't express, but they're growing fast and creating a two-tier stdlib.

### Fix (incremental, not blocking)
1. Audit each helper function: can any be expressed in `.kuki` now? The accessor functions (`EventID`, `PodEventType`, etc.) are already in `.kuki` — good.
2. For functions that must stay in Go, add a `// kukicha:helper` comment convention so they're easy to grep and track.
3. Track the ratio of `.kuki` vs `.go`-only code in each stdlib package. If a package is >50% raw Go, consider whether the SDK it wraps is too complex for the stdlib.

No immediate code changes — this is a process/tracking item.

---

## P4 - Minor: obs.Log stdout-only

**File:** `stdlib/obs/obs.kuki`

### Problem
`obs.Log` writes to `fmt.Printf` (stdout) with no way to redirect to stderr, a file, or a custom writer. This is fine for scripts but limits production use.

### Fix (defer until needed)
When a user requests it, add an optional `io.Writer` field to `Logger`:

```kukicha
type Logger
    service string
    environment string
    component string
    correlationID string
    writer io.Writer        # defaults to os.Stdout
```

No immediate code change needed — the current API is correct for the target audience.
