# Kukicha Language Support for VS Code

Full language support for [Kukicha](https://github.com/duber000/kukicha) - a beginner-friendly programming language that transpiles to Go.

## Features

- 🎨 **Syntax Highlighting** - Rich syntax highlighting for `.kuki` files
- 🚀 **LSP Integration** - Full Language Server Protocol support including:
  - Go to Definition
  - Hover information
  - Auto-completion
  - Real-time diagnostics
  - Find references
- ⚡ **Fast Performance** - Pre-compiled LSP binary bundled with the extension
- 🔧 **Format on Save** - Automatic code formatting (configurable)

## Installation

### From VS Code Marketplace

1. Open VS Code
2. Press `Ctrl+P` (or `Cmd+P` on macOS)
3. Type `ext install kukicha-lang`
4. Press Enter

### From VSIX

1. Download the latest `.vsix` file from [GitHub Releases](https://github.com/duber000/kukicha/releases)
2. In VS Code, open the Command Palette (`Ctrl+Shift+P` / `Cmd+Shift+P`)
3. Select "Extensions: Install from VSIX"
4. Choose the downloaded `.vsix` file

## Usage

Once installed, the extension automatically activates when you open a `.kuki` file.

### Keyboard Shortcuts

- **Format Document**: `Shift+Alt+F` (Windows/Linux) or `Shift+Option+F` (macOS)
- **Go to Definition**: `F12`
- **Peek Definition**: `Alt+F12`
- **Find References**: `Shift+F12`
- **Hover**: Hover over any identifier with your mouse

### Commands

The extension provides the following commands (available via Command Palette):

- `Kukicha: Format Kukicha File` - Format the current file
- `Kukicha: Check Syntax` - Run syntax check on the current file

## Configuration

Add these settings to your `settings.json`:

```json
{
  // Custom path to kukicha-lsp binary (leave empty to use bundled binary)
  "kukicha.lsp.path": "",
  
  // Enable automatic formatting on save
  "kukicha.format.enable": true,
  
  // Trace LSP communication (for debugging)
  "kukicha.trace.server": "off"
}
```

Valid values for `kukicha.trace.server`:
- `"off"` - No tracing
- `"messages"` - Log requests and responses
- `"verbose"` - Log all communication

## Development

### Prerequisites

- Node.js 18+
- npm
- Go 1.21+ (for building LSP binary)

### Setup

```bash
cd editors/vscode
npm install
```

### Build

```bash
# Build the extension
npm run esbuild

# Build and watch for changes
npm run esbuild-watch
```

### Run LSP Binary Builder

```bash
# Build LSP binaries for all platforms (requires Go)
./scripts/build-lsp.sh

# Or fetch pre-built binaries from GitHub releases
npm run fetch-lsp
```

### Debug

1. Open this directory in VS Code
2. Copy `.vscode/launch.example.json` to `.vscode/launch.json`
3. Press `F5` to launch the Extension Development Host
4. Open a `.kuki` file to test

### Package

```bash
npm install -g @vscode/vsce
vsce package
```

### Publish

```bash
vsce publish
```

## Features in Detail

### Syntax Highlighting

The extension includes a comprehensive TextMate grammar that highlights:

- Keywords (`func`, `type`, `if`, `for`, `switch`, etc.)
- Types (primitives, generics like `list of`, `map of ... to`)
- String interpolation (`"Hello {name}!"`)
- Pipe operators (`|>`)
- Arrow lambdas (`=>`)
- Error handling (`onerr`)
- And more!

### Language Server

The bundled LSP server provides:

- **Hover**: Type information and documentation on hover
- **Go to Definition**: Jump to variable, function, or type definitions
- **Completions**: Intelligent auto-completion suggestions
- **Diagnostics**: Real-time error reporting
- **Find References**: Find all usages of a symbol

## Troubleshooting

### LSP Server Not Starting

1. Check the Output panel (View → Output) and select "Kukicha Language Server"
2. Verify the bundled binary exists in `binaries/` directory
3. Try setting a custom path: `"kukicha.lsp.path": "/path/to/kukicha-lsp"`

### Syntax Highlighting Not Working

1. Check that the file has `.kuki` extension
2. Verify the language mode is set to "Kukicha" (bottom right of VS Code)
3. Reload VS Code window: `Ctrl+Shift+P` → "Developer: Reload Window"

### Formatting Not Working

1. Ensure `kukicha.format.enable` is set to `true`
2. Check that the `kukicha` compiler is installed and in your PATH
3. Look for error messages in the Output panel

## Known Issues

- Tree-sitter-WASM syntax highlighting is experimental. The TextMate grammar is used as the primary highlighter.
- Windows users may need to restart VS Code after installation for the LSP binary to be recognized.

## Contributing

Contributions are welcome! Please open an issue or submit a PR on the [main Kukicha repository](https://github.com/duber000/kukicha).

## License

Apache License 2.0 - See the [LICENSE](../../LICENSE) file for details.

## Acknowledgments

- Built with [vscode-languageclient](https://github.com/microsoft/vscode-languageserver-node)
- Syntax highlighting inspired by the [Zed extension](../zed/)
- LSP server: [kukicha-lsp](../../cmd/kukicha-lsp/)
