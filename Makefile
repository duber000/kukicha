# Kukicha build system
#
# Kukicha requires Go 1.25+ and uses GOEXPERIMENT=jsonv2,greenteagc.
# All make targets set this automatically. If running go commands
# directly, export GOEXPERIMENT=jsonv2,greenteagc first.
#
# The stdlib/*.go files are generated from stdlib/*.kuki sources.
# Always edit the .kuki files, then run `make generate` to update.

export GOEXPERIMENT := jsonv2,greenteagc

KUKICHA := ./kukicha
KUKI_SOURCES := $(wildcard stdlib/*/*.kuki)
# Exclude test files from generation
KUKI_MAIN := $(filter-out %_test.kuki,$(KUKI_SOURCES))

.PHONY: all build generate test check-generate clean

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
