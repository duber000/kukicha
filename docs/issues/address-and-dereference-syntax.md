# Language Feature: Address-of and Dereference Syntax

**Status:** Proposed
**Priority:** High - Required for pure Kukicha stdlib usage
**Component:** Lexer, Parser, Codegen

---

## Problem

Kukicha currently has no pure syntax for:
1. Taking the address of an existing variable (Go's `&variable`)
2. Explicitly dereferencing a pointer (Go's `*pointer`)

This forces users to drop into Go syntax when using Go stdlib functions that require pointers:

```kukicha
# Currently impossible in pure Kukicha
json.Unmarshal(data, &user)  # & is rejected by lexer
```

The lexer actively rejects `&` with: "Unexpected character '&'. Did you mean '&&'?"

---

## Proposed Solution

Add two new syntactic constructs that follow Kukicha's English-like patterns:

### 1. Address-of: `reference of`

```kukicha
# Taking address of existing variable
json.Unmarshal(data, reference of user)

# Returning a pointer to local variable
return reference of result
```

**Compiles to Go:**
```go
json.Unmarshal(data, &user)
return &result
```

### 2. Dereference: `dereference`

```kukicha
# Get value from pointer
user := dereference userPtr

# Assign through pointer
dereference ptr = newValue
```

**Compiles to Go:**
```go
user := *userPtr
*ptr = newValue
```

---

## Syntax Patterns

This follows Kukicha's existing "of" patterns:

| Pattern | Example | Go Equivalent |
|---------|---------|---------------|
| `list of Type` | `list of int` | `[]int` |
| `map of K to V` | `map of string to int` | `map[string]int` |
| `reference Type` | `reference User` | `*User` |
| `reference to Struct {}` | `reference to Node { value: 1 }` | `&Node{Value: 1}` |
| **`reference of var`** | `reference of user` | `&user` |
| **`dereference ptr`** | `dereference userPtr` | `*userPtr` |

---

## Use Cases

### JSON Unmarshaling

```kukicha
import "encoding/json"

func LoadConfig(data list of byte) (Config, error)
    config := Config{}
    json.Unmarshal(data, reference of config) onerr return Config{}, error
    return config, nil
```

### Database Scanning

```kukicha
import "database/sql"

func GetUser(db reference sql.DB, id int64) (User, error)
    user := User{}
    row := db.QueryRow("SELECT name, email FROM users WHERE id = ?", id)
    row.Scan(reference of user.Name, reference of user.Email) onerr return User{}, error
    return user, nil
```

### Linked List Operations

```kukicha
type Node
    value int
    next reference Node

func Append(head reference Node, value int)
    current := head
    for current.next not equals empty
        current = current.next
    current.next = reference to Node
        value: value
        next: empty

# Explicit dereference when needed
func GetValue(ptr reference Node) int
    return dereference ptr .value  # or just ptr.value with auto-deref
```

### Swap Function

```kukicha
func Swap(a reference int, b reference int)
    temp := dereference a
    dereference a = dereference b
    dereference b = temp
```

---

## Implementation

### Lexer Changes

**File:** `internal/lexer/lexer.go`

1. Keep `&` rejection for now (or make it emit a helpful error pointing to `reference of`)
2. No new tokens needed - `reference`, `of`, and `dereference` are already keywords or can be added

**File:** `internal/lexer/token.go`

Add if not present:
```go
TOKEN_DEREFERENCE  // dereference
```

### Parser Changes

**File:** `internal/parser/parser.go`

1. In `parsePrimaryExpr` or `parseUnaryExpr`:
   - Detect `reference of` followed by expression → `AddressOfExpr`
   - Detect `dereference` followed by expression → `DerefExpr`

2. In assignment parsing:
   - Allow `dereference expr = value` as assignment target

### AST Changes

**File:** `internal/ast/ast.go`

```go
// AddressOfExpr represents "reference of variable"
type AddressOfExpr struct {
    Token   lexer.Token  // the 'reference' token
    Operand Expression   // the variable to take address of
}

// DerefExpr represents "dereference pointer"
type DerefExpr struct {
    Token   lexer.Token  // the 'dereference' token
    Operand Expression   // the pointer to dereference
}
```

### Codegen Changes

**File:** `internal/codegen/codegen.go`

```go
func (g *Generator) generateAddressOfExpr(expr *ast.AddressOfExpr) string {
    return "&" + g.generateExpr(expr.Operand)
}

func (g *Generator) generateDerefExpr(expr *ast.DerefExpr) string {
    return "*" + g.generateExpr(expr.Operand)
}
```

---

## Migration

Once implemented, update documentation to show pure Kukicha patterns:

**Before (mixed syntax):**
```kukicha
json.Unmarshal(data, &user)  # Go syntax leak
```

**After (pure Kukicha):**
```kukicha
json.Unmarshal(data, reference of user)
```

---

## Testing

### Parser Tests

```go
func TestAddressOfExpr(t *testing.T) {
    input := `x := reference of user`
    // Should parse as VarDecl with AddressOfExpr
}

func TestDerefExpr(t *testing.T) {
    input := `value := dereference ptr`
    // Should parse as VarDecl with DerefExpr
}

func TestDerefAssignment(t *testing.T) {
    input := `dereference ptr = newValue`
    // Should parse as assignment with DerefExpr on left
}
```

### Codegen Tests

```go
func TestAddressOfCodegen(t *testing.T) {
    input := `json.Unmarshal(data, reference of user)`
    expected := `json.Unmarshal(data, &user)`
}
```

---

## Open Questions

1. **Precedence:** How does `reference of` bind? Is `reference of user.field` the address of `user.field` or `(reference of user).field`?
   - Proposal: `reference of` binds tightly to the immediately following expression

2. **Chaining:** Should `reference of reference of x` work (pointer to pointer)?
   - Proposal: Yes, for completeness

3. **Error messages:** When user writes `&x`, should the error say "use 'reference of x' instead"?
   - Proposal: Yes, for discoverability

---

## Related

- [tuple-return-limitation.md](tuple-return-limitation.md) - Closed, but this feature enables full stdlib usage
- [kukicha-design-philosophy.md](../kukicha-design-philosophy.md) - Update after implementation

---

**Created:** 2026-01-22
**Author:** Design discussion
