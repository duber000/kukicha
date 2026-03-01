# Changelog

All notable changes to the Kukicha VS Code extension will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial release of Kukicha VS Code extension
- Syntax highlighting for `.kuki` files via TextMate grammar
- Full LSP integration (hover, go-to-definition, completions, diagnostics)
- Pre-bundled LSP binaries for Linux, macOS (x64/ARM64), and Windows
- Format on save support
- Custom commands: Format Kukicha File, Check Syntax
- Configurable LSP path for custom installations
- Extension icon and branding

## [0.1.0] - 2026-02-28

### Added
- Initial public release
- Based on Kukicha compiler v0.0.11
- TextMate grammar with comprehensive syntax highlighting
- Language server protocol client integration
- Cross-platform LSP binary bundling
- Documentation and examples

### Features Included
- Hover information for types, functions, and variables
- Go-to-definition navigation
- Auto-completion suggestions
- Real-time error diagnostics
- Code folding support
- Auto-closing brackets and quotes
- 4-space indentation enforcement

### Known Issues
- Tree-sitter-WASM syntax highlighting not yet integrated (using TextMate grammar)
- Windows users may need to restart VS Code after installation

### Planned
- Tree-sitter-WASM integration for enhanced syntax highlighting
- Code formatting via `kukicha fmt` integration
- Snippets for common patterns
- Debug adapter protocol (DAP) integration
- Integrated test runner
