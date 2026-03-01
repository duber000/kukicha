#!/bin/bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
EXTENSION_DIR="$(dirname "$SCRIPT_DIR")"
PROJECT_ROOT="$(dirname "$(dirname "$EXTENSION_DIR")")"
BINARIES_DIR="$EXTENSION_DIR/binaries"

echo "Building kukicha-lsp binaries for all platforms..."

mkdir -p "$BINARIES_DIR"

cd "$PROJECT_ROOT"

echo "Building for Linux (x86_64)..."
GOOS=linux GOARCH=amd64 go build -o "$BINARIES_DIR/linux/kukicha-lsp" ./cmd/kukicha-lsp

echo "Building for macOS (x86_64)..."
GOOS=darwin GOARCH=amd64 go build -o "$BINARIES_DIR/darwin-x64/kukicha-lsp" ./cmd/kukicha-lsp

echo "Building for macOS (ARM64)..."
GOOS=darwin GOARCH=arm64 go build -o "$BINARIES_DIR/darwin-arm64/kukicha-lsp" ./cmd/kukicha-lsp

echo "Building for Windows (x86_64)..."
GOOS=windows GOARCH=amd64 go build -o "$BINARIES_DIR/win32/kukicha-lsp.exe" ./cmd/kukicha-lsp

echo "Done! Binaries are in $BINARIES_DIR/"
ls -la "$BINARIES_DIR/"*/

