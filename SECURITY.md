# Kukicha Security Audit & Roadmap

Security findings and planned mitigations for the Kukicha language and standard library.

## High Severity

### 1. String interpolation bypasses SQL parameterization

**Status:** ✅ Fixed — see Completed section

Kukicha's `"text {variable}"` syntax interpolates before the string reaches pgx,
silently defeating parameterized queries:

```kukicha
# UNSAFE — interpolation happens before pgx sees the string
pg.Query(pool, "SELECT * FROM {table} WHERE id = $1", id)

# SAFE — use $1 parameters for all dynamic values
pg.Query(pool, "SELECT * FROM users WHERE id = $1", id)
```

**Mitigation:** Compiler error when string interpolation is detected in the first
string argument to `pg.Query`, `pg.QueryRow`, `pg.Exec`, and their `Tx` variants.

### 2. `http.HTML()` writes raw unescaped content

**Status:** ✅ Fixed

`http.HTML(w, content)` writes `content` verbatim with `text/html` content type.
Any user input passed here is a direct XSS vector.

**Mitigations applied:**
- Added `http.SafeHTML(w, content)` that HTML-escapes content via `html.EscapeString`
- Compiler error when `http.HTML` (or its `httphelper` alias) receives a non-literal
  argument: `XSS risk: http.HTML with non-literal content — use http.SafeHTML`

### 3. `template` uses `text/template` (no HTML escaping)

**Status:** ✅ Fixed

Go's `text/template` performs zero HTML escaping. Using this package to generate
HTML responses is unsafe.

**Mitigations applied:**
- Added `template.HTMLExecute` using `html/template` (auto-escapes `{{ }}` values)
- Added `template.HTMLRenderSimple` using `html/template` for one-call rendering
- `template.Execute` doc comment now warns: "For plaintext output only. Use HTMLExecute for HTML."

### 4. `fetch` has no SSRF protection by default

**Status:** ✅ Fixed

`fetch.Get(url)` will fetch internal IPs, cloud metadata endpoints, etc.
The `netguard` package is excellent but entirely opt-in.

**Mitigations applied:**
- Added `fetch.SafeGet(url)` that wraps `netguard.NewSSRFGuard()` automatically
- Compiler error when `fetch.Get`, `fetch.Post`, or `fetch.New` is used inside an
  HTTP handler: `SSRF risk: fetch.Get inside an HTTP handler — use fetch.SafeGet`

### 5. `files.*` has no path traversal protection

**Status:** ✅ Fixed

`files.Read(userInput)` reads any file the process can access. The safe
`sandbox.*` package exists but nothing guides developers toward it.

**Mitigations applied:**
- Compiler error when `files.Read`, `files.Write`, `files.Delete`, and other I/O
  functions are called inside an HTTP handler: `path traversal risk: files.Read
  inside an HTTP handler — use sandbox.* with a restricted root`

## Medium Severity

### 6. `shell.Run` with non-literal strings

**Status:** 🟡 Open

`shell.Run(cmd)` splits a single string on whitespace. The doc says "literals only"
but nothing enforces this.

**Mitigation:** Compiler warning when `shell.Run` argument is not a string literal.

### 7. `http.Redirect` accepts unvalidated URLs

**Status:** 🟡 Open

`http.Redirect(w, r, url)` with user-controlled `url` is an open redirect.

**Mitigation:**
- Add `http.SafeRedirect(w, r, url, allowedHosts)`
- Compiler warning when `http.Redirect` receives a non-literal URL

### 8. No HTTP response body size limits

**Status:** 🟡 Open

`fetch.Text()`, `fetch.Bytes()`, `json.UnmarshalRead()`, `http.ReadJSON()` all
use unbounded `io.ReadAll`, enabling OOM/DoS.

**Mitigation:**
- Add `fetch.MaxBodySize()` builder option using `io.LimitReader`
- Add `http.ReadJSONLimit(w, r, maxBytes, &target)`
- Consider default limits with opt-out

### 9. No security headers from `http.Serve()`

**Status:** 🟡 Open

No `X-Content-Type-Options`, `X-Frame-Options`, `Strict-Transport-Security`,
`Content-Security-Policy`, or `Referrer-Policy` headers are set.

**Mitigation:** Add `http.SecureHeaders()` middleware that sets sensible defaults.

### 10. No security-related static analysis in compiler

**Status:** 🟡 Open — SQL interpolation check in progress

The compiler performs zero security checks. No taint tracking, no dangerous-pattern
warnings, no lint-style security checks.

**Mitigation:** Incremental addition of checks (this document tracks each one).

## Low Severity

### 11. `validate` lacks security-specific sanitizers

**Status:** 🟢 Open

No `NoHTML`, `SafeFilename`, or `NoNullBytes` validators exist.

**Mitigation:** Add security-focused validators to `stdlib/validate`.

### 12. `validate.Matches` compiles regex on every call

**Status:** 🟢 Open

Under adversarial input, a user-supplied regex pattern could cause ReDoS.

**Mitigation:** Pre-compile regex or add `validate.MatchesCompiled`.

### 13. `pg.Connect` encourages plaintext credentials

**Status:** 🟢 Open

Examples show literal connection strings with passwords in source code.

**Mitigation:** Update examples to use `env.Get("DATABASE_URL")` pattern.

## Completed

### 1. String interpolation bypasses SQL parameterization

**Status:** ✅ Fixed — compiler rejects `"{var}"` in pg.Query/QueryRow/Exec/Tx\* SQL arguments

### 2. `http.HTML()` writes raw unescaped content

**Status:** ✅ Fixed — `http.SafeHTML` added; compiler warns on non-literal arg to `http.HTML`

### 3. `template` uses `text/template` (no HTML escaping)

**Status:** ✅ Fixed — `template.HTMLExecute` and `template.HTMLRenderSimple` added using `html/template`

### 4. `fetch` has no SSRF protection by default

**Status:** ✅ Fixed — `fetch.SafeGet` added; compiler warns when `fetch.Get/Post/New` used inside HTTP handler

### 5. `files.*` has no path traversal protection

**Status:** ✅ Fixed — compiler warns when `files.*` I/O functions are called inside HTTP handlers
