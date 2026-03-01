
# Kukicha VSCode Extension Development Guide

This directory contains the VSCode extension for Kukicha language support.

## Quick Start

```bash
# Install dependencies
npm install

# Build extension
npm run esbuild

# Build with watch mode
npm run esbuild-watch

# Fetch LSP binaries (alternative to building from source)
npm run fetch-lsp
```

## LSP Binary Options

You have two options for getting the LSP server:

### Option 1: Fetch Pre-built Binaries (Recommended)

```bash
npm run fetch-lsp
```

This downloads pre-built binaries from GitHub releases for all platforms.

### Option 2: Build from Source

```bash
./scripts/build-lsp.sh
```

This requires Go 1.21+ to be installed and will cross-compile for all platforms.

## Directory Structure

```
vscode/
├── src/
│   └── extension.ts          # Main extension entry point
├── syntaxes/
│   └── kukicha.tmLanguage.json  # TextMate grammar for syntax highlighting
├── languages/
│   └── kukicha.json          # Language configuration
├── binaries/
│   ├── linux/kukicha-lsp     # Linux binary
│   ├── darwin-x64/kukicha-lsp # macOS Intel binary
│   ├── darwin-arm64/kukicha-lsp # macOS Apple Silicon binary
│   └── win32/kukicha-lsp.exe # Windows binary
├── scripts/
│   ├── build-lsp.sh          # Build script for LSP binaries
│   └── fetch-lsp-binaries.js # Download script for pre-built binaries
├── images/
│   ├── icon.png              # Extension icon
│   ├── icon-dark.svg         # Dark theme icon
│   └── icon-light.svg        # Light theme icon
├── package.json              # Extension manifest
├── tsconfig.json             # TypeScript configuration
└── README.md                 # User documentation (this file)
```

## Debugging

1. Open this directory in VS Code
2. Create `.vscode/launch.json` from the example:
   ```bash
   cp .vscode/launch.example.json .vscode/launch.json
   ```
3. Set breakpoints in `src/extension.ts`
4. Press `F5` to launch the Extension Development Host
5. Open a `.kuki` file in the new window

## Testing

Test the extension with different scenarios:

1. **Syntax Highlighting**: Open various `.kuki` files and verify highlighting
2. **LSP Features**:
   - Hover over identifiers
   - Try go-to-definition (F12)
   - Trigger completions (Ctrl+Space)
   - Introduce errors and check diagnostics
3. **Formatting**: Use format document command

## Publishing

```bash
# Install vsce if not already installed
npm install -g @vscode/vsce

# Login to Microsoft account
vsce login kukicha

# Package the extension
vsce package

# Publish to marketplace
vsce publish
```

## Versioning

Update the version in `package.json` following semantic versioning:
- MAJOR.MINOR.PATCH (e.g., 0.1.0)

Also update the `KUKICHA_VERSION` in `scripts/fetch-lsp-binaries.js` to match the compiler version.

## Common Issues

### Binary Not Found

If the LSP binary isn't found at runtime:
1. Verify it exists in the appropriate `binaries/` subdirectory
2. Check file permissions (should be executable)
3. On Windows, ensure the `.exe` extension is present

### Syntax Highlighting Issues

If syntax highlighting breaks:
1. Check the grammar file is valid JSON
2. Reload the VS Code window
3. Check the Developer Tools console for errors

### LSP Not Starting

Check the Output panel → "Kukicha Language Server" for error messages.

## Resources

- [VS Code Extension API Documentation](https://code.visualstudio.com/api)
- [Language Server Protocol Specification](https://microsoft.github.io/language-server-protocol/)
- [vscode-languageclient Documentation](https://github.com/microsoft/vscode-languageserver-node)
- [Main Kukicha Repository](https://github.com/duber000/kukicha)
