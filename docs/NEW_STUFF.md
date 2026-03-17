# New Features & Expansion Plan

Roadmap for new syntax, stdlib packages, and real-world examples.

---

## Phase 1: New Syntax

Foundational language features that unblock everything else.

### 1. Const Blocks

**Status:** Not implemented
**Impact:** High | **Effort:** Medium
**QUESTION:** constant alias for new users like function and variable?

Every real program needs constants. Currently there's no way to express them in Kukicha.

**Proposed syntax:**
```kukicha
const MaxRetries = 5
const DefaultPort = 8080
const AppName = "myservice"

# Grouped form
const
    StatusOK = 200
    StatusNotFound = 404
    StatusServerError = 500
```

**Unlocks:** Enum-like patterns, config constants, HTTP status codes, protocol constants.

**Compiler work:** New `TOKEN_CONST` keyword, parser rule for `const` and `const` blocks (indented group), AST `ConstDecl` node, codegen to Go `const` declarations.

### 2. Multi-Line Strings

**Status:** Not implemented
**Impact:** High | **Effort:** Medium
**QUESTION:** More new symbols for new users to memorize, what do Go, Python, Ruby and Julia do here?

SQL queries, JSON templates, help text, and embedded config are all painful with single-line strings today. Critical for REST API and data pipeline examples.

**Proposed syntax (triple-quote):**
```kukicha
query := """
    SELECT u.name, u.email
    FROM users u
    WHERE u.active = $1
    ORDER BY u.created_at DESC
    """

helpText := """
    Usage: mytool [command]

    Commands:
      list     List all items
      add      Add a new item
      remove   Remove an item
    """
```

**Behavior:**
- Content between `"""` delimiters
- Leading common whitespace stripped (dedent)
- String interpolation works inside: `"... {expr} ..."`
- Raw variant (no interpolation) TBD — could use backticks or `r"""`

**Compiler work:** Lexer support for `"""` delimiter, dedent logic, integration with existing interpolation tokenization (`TOKEN_STRING_HEAD`/`MID`/`TAIL`).

### 3. Guard Clauses

**Status:** Not implemented
**Impact:** Low-Medium | **Effort:** Low
**Question:** How can we make this readable for new users?

Syntactic sugar for early-return validation. Reduces nesting in functions with multiple preconditions.

**Proposed syntax:**
```kukicha
func CreateUser(name string, age int) (User, error)
    guard name != "" else return User{}, error "name required"
    guard age >= 0 else return User{}, error "age must be non-negative"

    # ... main logic with no nesting
```

**Equivalent to:**
```kukicha
    if name == ""
        return User{}, error "name required"
    if age < 0
        return User{}, error "age must be non-negative"
```

**Compiler work:** New `TOKEN_GUARD` and `TOKEN_ELSE` keywords (or reuse `else`), parser rule, AST node, codegen to `if !cond { ... }`.

**Note:** Lower priority — the `if`/`return` pattern already works. Guard is a readability improvement, not a capability unlock.

---

## Phase 2: Stdlib Expansion

New packages that fill common gaps. Build after Phase 1 syntax is available.

### stdlib/crypto

**Purpose:** Hashing, HMAC, token generation, password hashing.
**Question:** What does this unlock in the examples? I'm not sure we need thisyet 

**Functions:**
| Function | Description |
|----------|-------------|
| `SHA256(data string) string` | Hex-encoded SHA-256 hash |
| `SHA256Bytes(data list of byte) list of byte` | Raw SHA-256 hash |
| `HMAC(key string, data string) string` | HMAC-SHA256, hex-encoded |
| `RandomToken(length int) string` | Crypto-random hex token |
| `RandomBytes(n int) list of byte` | Crypto-random bytes |
| `HashPassword(password string) string, error` | bcrypt hash |
| `CheckPassword(hash string, password string) error` | bcrypt verify |

**Wraps:** `crypto/sha256`, `crypto/hmac`, `crypto/rand`, `golang.org/x/crypto/bcrypt`

### stdlib/table

**Purpose:** Pretty-printed tables for CLI output.

**Functions:**
| Function | Description |
|----------|-------------|
| `New(headers list of string) Table` | Create table with headers |
| `AddRow(t Table, row list of string) Table` | Append a row |
| `Print(t Table)` | Print with auto-aligned columns |
| `PrintWithStyle(t Table, style string)` | Print with border style (`plain`, `box`, `markdown`) |
| `ToString(t Table) string` | Render to string |

**Would improve:** `gh-semver-release list` output, any CLI tool with tabular data.

### stdlib/sort

**Purpose:** Sort slices by key function.

**Functions:**
| Function | Description |
|----------|-------------|
| `By(items list of any, less func(any, any) bool) list of any` | Sort by comparator |
| `ByKey(items list of any, key func(any) any2) list of any` | Sort by extracted key |
| `Strings(items list of string) list of string` | Sort strings |
| `Ints(items list of int) list of int` | Sort ints |
| `Reverse(items list of any, less func(any, any) bool) list of any` | Sort descending |

**Wraps:** `slices.SortFunc`, `slices.SortStableFunc`

---

## Phase 3: Real-World Examples

Showcase examples that use the full language. Build after Phases 1-2.

### REST API Server

A complete CRUD API demonstrating production patterns.

**Showcases:** http, pg, json, validate, crypto, error handling, const blocks, multi-line strings, security checks

**Structure:**
```
examples/rest-api/
    main.kuki          # Routes, middleware, server startup
    handlers.kuki      # Request handlers
    models.kuki        # Types + validation
    db.kuki            # Database queries (multi-line SQL)
```

**Key patterns:**
- Const blocks for HTTP status codes and config
- Multi-line strings for SQL queries
- `onerr` for error propagation through handler chains
- Security: `SafeHTML`, parameterized queries, `ReadJSONLimit`, `SecureHeaders`
- `must.Env` for startup config, `env.Get` for runtime config

### Data Pipeline

A CLI tool that reads, transforms, and outputs data.

**Showcases:** files, parse, iterator, slice, concurrent, cli, table

**Structure:**
```
examples/data-pipeline/
    main.kuki          # CLI entry + pipeline orchestration
```

**Key patterns:**
- `iterator.Values |> iterator.Filter |> iterator.Map |> iterator.Collect` chains
- `concurrent.Parallel` for fan-out processing
- `table.Print` for formatted output
- `parse.CsvWithHeader` for input, `json.Marshal` for output
- Multi-line strings for format templates

### Deployment Tool

An interactive CLI for managing deployments.

**Showcases:** cli, container or kube, shell, obs, retry, input, semver

**Structure:**
```
examples/deploy-tool/
    main.kuki          # CLI subcommands + deployment logic
```

**Key patterns:**
- `cli.Command` subcommands: `deploy`, `rollback`, `status`
- `obs.Info`/`obs.Error` for structured logging with correlation IDs
- `retry.New` for transient failure handling
- `input.Confirm` for dangerous operations
- `semver.Parse`/`semver.Bump` for version management
- Const blocks for deployment config

---

## Execution Order

| Step | What | Depends On |
|------|------|------------|
| 1a | Const blocks | — |
| 1b | Multi-line strings | — |
| 1c | Guard clauses (optional) | — |
| 2a | stdlib/crypto | — |
| 2b | stdlib/table | — |
| 2c | stdlib/sort | — |
| 3a | REST API example | const blocks, multi-line strings, crypto |
| 3b | Data pipeline example | table, multi-line strings |
| 3c | Deployment tool example | const blocks |

Steps 1a and 1b can be done in parallel. Steps 2a-2c can be done in parallel. Step 3 examples depend on their prerequisites.
