.PHONY: install build watch package publish clean fetch-lsp build-lsp test

install:
	npm install

build:
	npm run esbuild

watch:
	npm run esbuild-watch

fetch-lsp:
	npm run fetch-lsp

build-lsp:
	./scripts/build-lsp.sh

package: build
	npx vsce package

publish: build
	npx vsce publish

clean:
	rm -rf dist node_modules *.vsix
	find binaries -type f ! -name '.gitkeep' -delete

test: build
	@echo "Extension build successful. Test by opening example.kuki in VS Code."
	@echo "Run 'npm run esbuild-watch' and press F5 to debug."
