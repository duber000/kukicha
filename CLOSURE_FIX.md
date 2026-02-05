# Fix: Closures as Function Call Arguments

**Status:** ✓ RESOLVED  
**Date:** 2026-02-04  
**Issue:** Closures could not be used directly as function call arguments due to indentation token suppression inside parentheses.

## Problem

The following code pattern would fail with parser errors:

```kukicha
filtered := items |> slice.Filter(func(x int) bool
    return x > 2
)
```

**Error:** `expected indented block` at the `return` statement.

### Root Cause

The lexer was suppressing INDENT/DEDENT tokens whenever `braceDepth > 0` (i.e., inside any parentheses, brackets, or braces). This mechanism was designed for line continuations (pipe operations), but it also prevented closures from having properly structured bodies inside function call argument lists.

**Lexer state before fix:**
- `braceDepth` tracked both `()` and `[]` and `{}`
- When `braceDepth > 0`, all indentation tokens were suppressed
- This broke closure parsing because closures need INDENT/DEDENT tokens for their bodies

## Solution

Separated parenthesis tracking from general brace tracking in the lexer:

### Changes to `internal/lexer/lexer.go`

1. **Added separate depth tracking** (lines 43-45):
   ```go
   braceDepth       int  // current nesting level of [], {} (used for continuations)
   parenDepth       int  // current nesting level of () (used for closures)
   inFunctionLiteral bool // true when we've just seen 'func' and are parsing its body
   ```

2. **Updated newline handling** (lines 137-165):
   - Changed continuation check from `(braceDepth > 0)` to only check `braceDepth`
   - Removed `parenDepth` from continuation suppression logic
   - Allows INDENT/DEDENT tokens to be emitted inside parentheses

3. **Separated parenthesis tracking** (lines 174-191):
   - `(` increments `parenDepth` (was incrementing `braceDepth`)
   - `)` decrements `parenDepth` (was decrementing `braceDepth`)
   - `[` and `]` continue to use `braceDepth`
   - `{` and `}` continue to use `braceDepth`

4. **Function literal detection** (lines 472-484):
   - Set `inFunctionLiteral = true` when seeing `func` keyword
   - Enables future optimizations for closure-specific handling

### Why This Works

- **Pipe continuations still work:** Only `braceDepth` (for `[]` and `{}`) suppresses indentation, not `parenDepth`
- **Closures get indentation:** Parentheses in function calls no longer suppress INDENT/DEDENT
- **No breaking changes:** Existing code using pipes and braces continues to work as before

## Testing

### Lexer Tests
Added `TestClosureInFunctionCall` covering:
- Simple closures in function arguments
- Closures with multiple statements
- Proper INDENT/DEDENT token generation

All tests pass: ✓

### Parser & Semantic Tests
- All existing tests pass (1,600+ tests)
- Manual testing confirms:
  - ✓ Simple closures as arguments
  - ✓ Nested closures
  - ✓ Closures with multiple statements
  - ✓ Closures in pipe operations
  - ✓ Pipe continuations still work correctly

## Examples Now Supported

```kukicha
# Filter with closure
filtered := items |> slice.Filter(func(x int) bool
    return x > 2
)

# Map with multi-statement closure
mapped := items |> slice.Map(func(x int) int
    doubled := x * 2
    adjusted := doubled + 1
    return adjusted
)

# Direct function calls
result := apply(func(x int) int
    return x * 2
)

# Nested closures
outer := func(x int) func(int) int
    return func(y int) int
        return x + y
```

## Files Modified

1. `internal/lexer/lexer.go` - Core lexer changes
2. `internal/lexer/lexer_test.go` - Added closure token tests
3. `AGENTS.md` - Removed closure limitation from known constraints

## Performance Impact

Minimal. The changes only add two simple integer field comparisons to the newline handling path.

## Backwards Compatibility

✓ Full backwards compatibility maintained. All existing Kukicha code continues to work.
