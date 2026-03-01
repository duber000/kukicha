#!/usr/bin/env bash

# Quick Start Script for Kukicha VS Code Extension Development
# This script helps you get set up and test the extension quickly

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "========================================="
echo "Kukicha VS Code Extension - Quick Start"
echo "========================================="
echo ""

cd "$SCRIPT_DIR"

# Check prerequisites
echo "Checking prerequisites..."

if ! command -v node &> /dev/null; then
    echo "❌ Node.js is not installed. Please install Node.js 18+ first."
    exit 1
fi

if ! command -v npm &> /dev/null; then
    echo "❌ npm is not installed. Please install npm first."
    exit 1
fi

if ! command -v go &> /dev/null; then
    echo "⚠️  Go is not installed. You'll need it to build LSP binaries."
    echo "   Alternatively, you can use 'npm run fetch-lsp' to download pre-built binaries."
fi

echo "✅ Node.js: $(node --version)"
echo "✅ npm: $(npm --version)"
if command -v go &> /dev/null; then
    echo "✅ Go: $(go version)"
fi
echo ""

# Install dependencies
echo "Installing npm dependencies..."
npm install --legacy-peer-deps
echo "✅ Dependencies installed"
echo ""

# Build the extension
echo "Building extension..."
npm run esbuild
echo "✅ Extension built successfully"
echo ""

# Check LSP binaries
echo "Checking LSP binaries..."
if [ -f "binaries/linux/kukicha-lsp" ] || [ -f "binaries/darwin-x64/kukicha-lsp" ] || [ -f "binaries/darwin-arm64/kukicha-lsp" ] || [ -f "binaries/win32/kukicha-lsp.exe" ]; then
    echo "✅ LSP binary found"
else
    echo "⚠️  No LSP binary found. You have two options:"
    echo ""
    echo "   Option 1: Download pre-built binaries (recommended)"
    echo "   Run: npm run fetch-lsp"
    echo ""
    echo "   Option 2: Build from source (requires Go)"
    echo "   Run: ./scripts/build-lsp.sh"
    echo ""
fi
echo ""

# Show next steps
echo "========================================="
echo "Next Steps:"
echo "========================================="
echo ""
echo "1. Debug the extension:"
echo "   - Copy .vscode/launch.example.json to .vscode/launch.json"
echo "   - Open this directory in VS Code"
echo "   - Press F5 to launch the Extension Development Host"
echo "   - Open example.kuki to test"
echo ""
echo "2. Test syntax highlighting:"
echo "   - Open example.kuki in VS Code"
echo "   - Verify keywords, types, and strings are highlighted"
echo ""
echo "3. Test LSP features:"
echo "   - Hover over identifiers to see type information"
echo "   - Press F12 to go to definition"
echo "   - Use Ctrl+Space for completions"
echo ""
echo "4. Package the extension:"
echo "   npm install -g @vscode/vsce"
echo "   vsce package"
echo ""
echo "========================================="
echo "Development server is ready!"
echo "Run 'npm run esbuild-watch' to auto-rebuild on changes"
echo "========================================="
