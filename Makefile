# Kukicha build system
#
# Kukicha requires Go 1.26+.
# The stdlib/*.go files are generated from stdlib/*.kuki sources.
# Always edit the .kuki files, then run `make generate` to update.

KUKICHA := ./kukicha
KUKI_SOURCES := $(wildcard stdlib/*/*.kuki)
# Exclude test files from generation
KUKI_MAIN := $(filter-out %_test.kuki,$(KUKI_SOURCES))

.PHONY: all build lsp generate genstdlibregistry test check-generate clean install-lsp

all: build lsp

# Build the kukicha compiler
build:
	go build -o $(KUKICHA) ./cmd/kukicha

# Regenerate internal/semantic/stdlib_registry_gen.go from stdlib/*.kuki signatures.
# Run this whenever a stdlib .kuki file adds, removes, or changes exported functions.
genstdlibregistry:
	go run ./cmd/genstdlibregistry

# Regenerate all stdlib .go files from .kuki sources.
# Runs genstdlibregistry first so the semantic registry stays in sync,
# then rebuilds the compiler with the fresh registry, then transpiles stdlib.
# Ignores go build errors (stdlib packages aren't standalone binaries).
generate: genstdlibregistry build
	@for f in $(KUKI_MAIN); do \
		echo "Transpiling $$f ..."; \
		out=$$($(KUKICHA) build "$$f" 2>&1); rc=$$?; \
		echo "$$out" | grep -v "^Warning: go build" || true; \
		if [ $$rc -ne 0 ]; then echo "ERROR: Failed to transpile $$f"; exit 1; fi; \
	done
	@echo "Done. Generated .go files from $(words $(KUKI_MAIN)) .kuki sources."

# Run all tests
test:
	go test ./...

# Check that generated .go files are up to date (for CI)
check-generate: generate
	@if [ -n "$$(git diff --name-only stdlib/ internal/semantic/stdlib_registry_gen.go)" ]; then \
		echo "ERROR: Generated files are out of date:"; \
		git diff --name-only stdlib/ internal/semantic/stdlib_registry_gen.go; \
		echo "Run 'make generate' and commit the results."; \
		exit 1; \
	fi
	@echo "Generated files are up to date."

clean:
	rm -f $(KUKICHA) ./kukicha-lsp

# Build the kukicha-lsp language server
lsp:
	go build -o ./kukicha-lsp ./cmd/kukicha-lsp

# Install the LSP server to GOPATH/bin (or ~/go/bin if GOPATH not set)
install-lsp: lsp
	cp ./kukicha-lsp $(shell go env GOPATH)/bin/
