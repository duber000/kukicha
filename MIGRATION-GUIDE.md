# Kukicha Migration Guide: v1.0 to v1.2

**Version:** 1.2.0
**Date:** 2026-01-20

---

## Overview

Kukicha v1.2 introduces two important language design improvements that enhance code readability and discoverability:

1. **`onerr` operator** - Replaces the overloaded `or` operator for error handling
2. **Explicit `this` receiver** - Makes method receivers visible in the signature

These changes align with Kukicha's core principle: **"readable with few basic rules"**

---

## Why These Changes?

### Problem 1: The `or` Operator Was Ambiguous

**Before (v1.0-1.1):**
```kukicha
# Is this boolean OR or error handling?
result := action() or "default"
```

You couldn't tell without knowing the return type of `action()`:
- If `action()` returns `bool` → Boolean OR
- If `action()` returns `(T, error)` → Error handling

**After (v1.2):**
```kukicha
# Error handling - crystal clear
result := action() onerr "default"

# Boolean OR - unchanged
if active or pending
    process()
```

✅ **Benefit**: Humans can distinguish at a glance without type knowledge

### Problem 2: The `this` Keyword Appeared Magically

**Before (v1.0-1.1):**
```kukicha
func Display on Todo string
    return this.title  # Where did 'this' come from?
```

Beginners asked: "Where did `this` come from? It's not in the parameters!"

**After (v1.2):**
```kukicha
func Display on this Todo string
    return this.title  # 'this' is right there in the signature!
```

✅ **Benefit**: The receiver is discoverable by reading the signature

---

## Breaking Changes

### Change 1: `or` → `onerr` for Error Handling

#### Old Syntax (v1.0-1.1)

```kukicha
# Panic on error
content := file.read("config.json") or panic "missing file"

# Return error
data := http.get(url) or return error "failed"

# Default value
port := env.get("PORT") or "8080"

# Multi-line handler
config := parseConfig(data) or
    print "Using defaults"
    loadDefaults()

# Chaining
config := file.read("config.json")
    or file.read("config.yaml")
    or panic "no config found"
```

#### New Syntax (v1.2)

```kukicha
# Panic on error
content := file.read("config.json") onerr panic "missing file"

# Return error
data := http.get(url) onerr return error "failed"

# Default value
port := env.get("PORT") onerr "8080"

# Multi-line handler
config := parseConfig(data) onerr
    print "Using defaults"
    loadDefaults()

# Chaining
config := file.read("config.json")
    onerr file.read("config.yaml")
    onerr panic "no config found"
```

#### Boolean `or` Remains Unchanged

```kukicha
# Boolean logic - no changes needed
if completed or archived
    markAsProcessed()

if active or pending or draft
    display()

while running or retrying
    attempt()
```

#### Migration Steps

**Option 1: Find and Replace (Simple)**

In your editor:
1. Find: ` or panic`
2. Replace: ` onerr panic`
3. Find: ` or return`
4. Replace: ` onerr return`
5. Find: ` or "`
6. Replace: ` onerr "`
7. Find: ` or empty`
8. Replace: ` onerr empty`

**Option 2: Automated Tool (Recommended)**

```bash
# Run the migration tool (coming soon)
kukicha migrate --from=1.1 --to=1.2 ./src/

# Dry run first (shows changes without applying)
kukicha migrate --from=1.1 --to=1.2 --dry-run ./src/
```

**Option 3: Regex (Advanced)**

```bash
# Find error handling 'or' patterns
grep -rn '\bor\s\+(panic\|return\|empty\|")' *.kuki

# Replace with sed (backup first!)
find . -name '*.kuki' -exec sed -i.bak 's/ or panic/ onerr panic/g' {} \;
find . -name '*.kuki' -exec sed -i.bak 's/ or return/ onerr return/g' {} \;
```

#### What NOT to Change

Do **not** change boolean `or` operators:

```kukicha
# These are CORRECT - do not change!
if active or pending
    process()

if condition1 or condition2 or condition3
    execute()

while running or retrying
    attempt()
```

### Change 2: Implicit `this` → Explicit `this`

#### Old Syntax (v1.0-1.1)

```kukicha
# Value receiver
func Display on Todo string
    return this.title

# Reference receiver
func MarkDone on reference Todo
    this.completed = true

# With parameters
func UpdateTitle on reference Todo, newTitle string
    this.title = newTitle
```

#### New Syntax (v1.2)

```kukicha
# Value receiver
func Display on this Todo string
    return this.title

# Reference receiver
func MarkDone on this reference Todo
    this.completed = true

# With parameters
func UpdateTitle on this reference Todo, newTitle string
    this.title = newTitle
```

#### Go-Style Syntax Unchanged

```kukicha
# Go syntax still works (no changes needed)
func (todo Todo) Display() string
    return todo.title

func (todo *Todo) MarkDone()
    todo.completed = true
```

#### Migration Steps

**Option 1: Find and Replace (Simple)**

In your editor:
1. Find: `func (\w+) on ([A-Z]\w+)`
2. Replace: `func $1 on this $2`
3. Find: `func (\w+) on reference ([A-Z]\w+)`
4. Replace: `func $1 on this reference $2`

**Option 2: Automated Tool (Recommended)**

```bash
# The migration tool handles this automatically
kukicha migrate --from=1.1 --to=1.2 ./src/
```

**Option 3: Manual (Small Codebases)**

For each method declaration:
1. Locate `func MethodName on Type`
2. Insert `this` after `on`: `func MethodName on this Type`
3. Same for reference receivers: `on reference Type` → `on this reference Type`

---

## Migration Checklist

Use this checklist to ensure complete migration:

### Error Handling (`or` → `onerr`)

- [ ] Update all `or panic` to `onerr panic`
- [ ] Update all `or return` to `onerr return`
- [ ] Update all `or "default"` to `onerr "default"`
- [ ] Update all `or empty` to `onerr empty`
- [ ] Update multi-line `or` handlers to `onerr`
- [ ] Verify boolean `or` operators remain unchanged
- [ ] Test error handling paths still work correctly

### Methods (Implicit → Explicit `this`)

- [ ] Update all `func Name on Type` to `func Name on this Type`
- [ ] Update all `func Name on reference Type` to `func Name on this reference Type`
- [ ] Verify Go-style methods (`func (t Type) Name()`) are unchanged
- [ ] Check that `this` is used consistently in method bodies
- [ ] Test all methods still compile and work correctly

### Testing & Validation

- [ ] Run `kukicha build` to check for syntax errors
- [ ] Run all unit tests
- [ ] Run integration tests
- [ ] Verify no behavioral changes
- [ ] Review git diff for unexpected changes

---

## Examples: Before & After

### Example 1: Error Handling

**Before (v1.1):**
```kukicha
func LoadConfig(path string) Config
    content := file.read(path) or return empty
    data := json.parse(content) or return empty
    return validateConfig(data) or return empty
```

**After (v1.2):**
```kukicha
func LoadConfig(path string) Config
    content := file.read(path) onerr return empty
    data := json.parse(content) onerr return empty
    return validateConfig(data) onerr return empty
```

### Example 2: Methods

**Before (v1.1):**
```kukicha
type Todo
    id int64
    title string
    completed bool

func Display on Todo string
    status := "pending"
    if this.completed
        status = "done"
    return "{status}: {this.title}"

func MarkDone on reference Todo
    this.completed = true
```

**After (v1.2):**
```kukicha
type Todo
    id int64
    title string
    completed bool

func Display on this Todo string
    status := "pending"
    if this.completed
        status = "done"
    return "{status}: {this.title}"

func MarkDone on this reference Todo
    this.completed = true
```

### Example 3: Complex Pipeline

**Before (v1.1):**
```kukicha
func GetRepoStats(username string) list of Repo
    return "https://api.github.com/users/{username}/repos"
        |> http.get()
        |> .json() as list of Repo
        |> filterByStars(10)
        |> sortByUpdated()
        or empty list of Repo
```

**After (v1.2):**
```kukicha
func GetRepoStats(username string) list of Repo
    return "https://api.github.com/users/{username}/repos"
        |> http.get()
        |> .json() as list of Repo
        |> filterByStars(10)
        |> sortByUpdated()
        onerr empty list of Repo
```

### Example 4: Full Application

**Before (v1.1):**
```kukicha
leaf todo

import time

type Todo
    id int64
    title string
    completed bool

func Display on Todo string
    return "{this.id}. {this.title}"

func Load(path string) list of Todo
    content := file.read(path) or return empty list of Todo
    todos := json.parse(content) or return empty list of Todo
    return todos
```

**After (v1.2):**
```kukicha
leaf todo

import time

type Todo
    id int64
    title string
    completed bool

func Display on this Todo string
    return "{this.id}. {this.title}"

func Load(path string) list of Todo
    content := file.read(path) onerr return empty list of Todo
    todos := json.parse(content) onerr return empty list of Todo
    return todos
```

---

## Timeline & Support

### Release Schedule

| Version | Status | Date | Description |
|---------|--------|------|-------------|
| v1.1.0 | Current | 2026-01-15 | Old syntax (deprecated) |
| v1.2.0-alpha | Released | 2026-01-20 | New syntax available |
| v1.2.0-beta | March 2026 | TBD | Deprecation warnings |
| v1.2.0 | May 2026 | TBD | Old syntax removed |

### Deprecation Warnings

Starting in **v1.2.0-beta** (March 2026), the compiler will emit warnings for old syntax:

```
Warning in todo.kuki:15:12
    content := file.read(path) or return empty
                                  ^^ The 'or' operator for error handling is deprecated.
                                     Use 'onerr' instead: content := file.read(path) onerr return empty

Help: Run 'kukicha migrate --from=1.1 --to=1.2 .' to automatically update your code.
```

### Support Policy

- **v1.1.x**: Supported until May 2026 (security fixes only)
- **v1.2.x**: Full support (new features, bug fixes)

---

## FAQ

### Q: Why break compatibility?

**A:** These changes address fundamental readability issues that would become harder to fix later. The migration is straightforward and can be largely automated.

### Q: Can I keep using the old syntax?

**A:** Only until v1.2.0 stable (May 2026). We strongly recommend migrating to prepare for the future.

### Q: Will the migration tool handle edge cases?

**A:** The tool handles 99% of cases. Complex macros or generated code may need manual review.

### Q: What if I use both Kukicha and Go style methods?

**A:** Only Kukicha-style methods (`on Type`) need updating. Go-style methods (`(t Type)`) are unchanged.

### Q: Does this affect compiled binaries?

**A:** No. The generated Go code is identical - only the Kukicha source syntax changes.

### Q: Can I migrate incrementally?

**A:** Yes! The changes are independent:
1. Migrate error handling (`or` → `onerr`) first
2. Then migrate methods (`on Type` → `on this Type`)

### Q: What about third-party libraries?

**A:** Library authors should migrate before v1.2.0 stable. The compiler will warn about outdated libraries.

---

## Getting Help

### Resources

- **Documentation**: [kukicha-syntax-v1.0.md](kukicha-syntax-v1.0.md)
- **Grammar**: [kukicha-grammar.ebnf.md](kukicha-grammar.ebnf.md)
- **Quick Reference**: [kukicha-quick-reference.md](kukicha-quick-reference.md)

### Support

- **GitHub Issues**: https://github.com/duber000/kukicha/issues
- **Discussions**: https://github.com/duber000/kukicha/discussions
- **Migration Tool**: `kukicha migrate --help`

### Reporting Migration Issues

If you encounter problems during migration:

1. Check this guide's FAQ section
2. Search existing GitHub issues
3. Create a new issue with:
   - Kukicha version (`kukicha version`)
   - Code snippet (before/after)
   - Error message
   - Expected vs actual behavior

---

## Appendix: Regex Patterns

For advanced users who want to create custom migration scripts:

### Error Handling Patterns

```regex
# Find error handling 'or' (not boolean)
(?<=\s)or\s+(panic|return|empty|")

# Boolean 'or' (should NOT be changed)
(?<=if\s+.*\s)or\s+
(?<=while\s+.*\s)or\s+
```

### Method Declaration Patterns

```regex
# Kukicha-style methods without 'this'
^(\s*)func\s+(\w+)\s+on\s+(reference\s+)?([A-Z]\w+)

# Replacement
\1func \2 on this \3\4
```

---

**Version:** 1.2.0
**Last Updated:** 2026-01-20
**Status:** Official Migration Guide

---

**Happy migrating! 🎉**

If you have questions or need help, don't hesitate to reach out via GitHub Issues or Discussions.
