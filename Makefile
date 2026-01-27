# Kukicha build system
#
# The stdlib/*.go files are generated from stdlib/*.kuki sources.
# Always edit the .kuki files, then run `make generate` to update.

KUKICHA := ./kukicha
KUKI_SOURCES := $(wildcard stdlib/*/*.kuki)
# Exclude test files from generation
KUKI_MAIN := $(filter-out %_test.kuki,$(KUKI_SOURCES))
GO_GENERATED := $(KUKI_MAIN:.kuki=.go)

.PHONY: all build generate test check clean

all: build

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
	rm -f $(KUKICHA)
