# Kukicha Standard Library - String Reference

This document contains additional string manipulation functions available in the `stdlib/string` package. These functions are not covered in the beginner tutorial but are useful for more advanced string operations.

## Core String Functions Used in Tutorials

The following functions are taught in the [Beginner Tutorial](tutorial/beginner-tutorial.md) and are essential:

- `string.ToLower()` - Convert to lowercase
- `string.Title()` - Convert to Title Case
- `string.TrimSpace()` - Remove leading/trailing whitespace
- `string.Split()` - Break a string into pieces
- `string.Join()` - Combine strings with a separator
- `string.Contains()` - Check if a string contains another string
- `string.HasPrefix()` - Check if a string starts with another string
- `string.HasSuffix()` - Check if a string ends with another string
- `string.Index()` - Find the position of a substring
- `string.TrimPrefix()` - Remove a prefix from a string
- `string.Count()` - Count occurrences of a substring
- `string.ReplaceAll()` - Replace all occurrences of a substring

Used in subsequent tutorials:
- `string.SplitN()` - Split with a maximum number of parts (console-todo)

## Additional String Functions

### Case Conversion

```kukicha
import "stdlib/string"

# Convert to UPPERCASE
upper := string.ToUpper("hello")  # Returns: "HELLO"
```

### Text Searching

```kukicha
import "stdlib/string"

text := "hello world"

# Find last occurrence (searches from the end)
lastIndex := string.LastIndex(text, "l")  # Returns: 9

# Find in lowercase version
searchText := string.ToLower("HELLO WORLD")
if "world" in searchText
    print("Found!")
```

### String Building

```kukicha
import "stdlib/string"

# Repeat a string
repeated := string.Repeat("ab", 3)  # Returns: "ababab"

# Trim specific characters from both ends
trimmed := string.Trim("  hello  ", " ")  # Returns: "hello"

# Trim from the left only
leftTrimmed := string.TrimLeft("  hello  ", " ")  # Returns: "hello  "

# Trim from the right only
rightTrimmed := string.TrimRight("  hello  ", " ")  # Returns: "  hello"

# Trim a specific suffix
noExt := string.TrimSuffix("file.txt", ".txt")  # Returns: "file"
```

### String Comparison

```kukicha
import "stdlib/string"

text1 := "hello"
text2 := "world"

# Case-insensitive comparison
if string.EqualFold(text1, "HELLO")
    print("Strings are equal ignoring case")

# Replace with case-insensitive matching
import "stdlib/string"
text := "The cat and the Cat"
# ReplaceAll is case-sensitive, use a loop with ToLower for case-insensitive
```

### Advanced Operations

```kukicha
import "stdlib/string"

# Split by multiple characters or patterns
parts := string.Fields("hello   world  from   kukicha")
# Returns: ["hello", "world", "from", "kukicha"]
# Useful for splitting on any whitespace

# Split with limit on number of parts
parts := string.SplitN("a:b:c:d", ":", 2)
# Returns: ["a", "b:c:d"]
```

## See Also

- [Beginner Tutorial](tutorial/beginner-tutorial.md) - Learn essential string operations
- [Kukicha Syntax Reference](kukicha-syntax-v1.0.md) - Complete language reference
- [Quick Reference](kukicha-quick-reference.md) - Quick lookup cheat sheet
