# Kukicha build system
#
# Kukicha requires Go 1.26+ and uses GOEXPERIMENT=jsonv2.
# All make targets set this automatically. If running go commands
# directly, export GOEXPERIMENT=jsonv2 first.
#
# The stdlib/*.go files are generated from stdlib/*.kuki sources.
# Always edit the .kuki files, then run `make generate` to update.

export GOEXPERIMENT := jsonv2

KUKICHA := ./kukicha
KUKI_SOURCES := $(wildcard stdlib/*/*.kuki)
# Exclude test files from generation
KUKI_MAIN := $(filter-out %_test.kuki,$(KUKI_SOURCES))

.PHONY: all build lsp generate test check-generate clean install-lsp

all: build lsp

# Build the kukicha compiler
build:
	go build -o $(KUKICHA) ./cmd/kukicha

# Regenerate all stdlib .go files from .kuki sources
# Ignores go build errors (stdlib packages aren't standalone binaries)
generate: build
	@for f in $(KUKI_MAIN); do \
		echo "Transpiling $$f ..."; \
		$(KUKICHA) build "$$f" 2>&1 | grep -v "^Warning: go build" || true; \
	done
	@echo "Done. Generated .go files from $(words $(KUKI_MAIN)) .kuki sources."

# Run all tests
test:
	go test ./...

# Check that generated .go files are up to date (for CI)
check-generate: generate
	@if [ -n "$$(git diff --name-only stdlib/)" ]; then \
		echo "ERROR: Generated .go files are out of date:"; \
		git diff --name-only stdlib/; \
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
