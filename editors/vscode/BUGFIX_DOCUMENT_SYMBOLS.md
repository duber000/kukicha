# Bug Fix: Document Symbols Error

## Issue
Error when requesting document symbols in VS Code:
```
Error: Illegal argument: character must be non-negative
```

## Root Cause
In `internal/lsp/completion.go`, the `getDocumentSymbols()` function was calculating positions by subtracting 1 from the column number returned by the AST's `Pos()` method:

```go
Character: d.Pos().Column - 1
```

When `d.Pos().Column` was 0, this resulted in -1, which is invalid for LSP positions (must be non-negative).

## Fix
Added a helper function `pos()` that ensures non-negative values:

```go
pos := func(line, col int) lsp.Position {
    return lsp.Position{
        Line:      max(0, line-1),
        Character: max(0, col-1),
    }
}
```

Applied this helper consistently throughout the function for all symbol types:
- Petiole declarations
- Function declarations
- Type declarations and fields
- Interface declarations and methods

## Testing
1. Rebuild with `make build` and `make lsp`
2. Rebuild binaries with `./scripts/build-lsp.sh`
3. Repackage extension with `vsce package`
4. Reinstall the updated .vsix file

## Files Changed
- `internal/lsp/completion.go` - Added position validation

## To Apply the Fix
1. Rebuild: `cd /var/home/tluker/repos/go/kukicha && make build && make lsp`
2. Update extension binaries: `cd editors/vscode && ./scripts/build-lsp.sh`
3. Rebuild extension: `vsce package --no-dependencies`
4. Reinstall the .vsix file in VS Code
