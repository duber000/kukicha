---
name: release
description: Cut a new Kukicha release — bump version, regenerate stdlib, run tests, commit, tag, and push. Use when the user says "cut a release", "release a new version", or similar.
---

# Release Skill

Cuts a new Kukicha release following the process defined in `docs/contributing.md`.

## Steps

### 1. Determine the next version

Read the current version from `internal/version/version.go`. Increment the patch segment (e.g. `0.0.10` → `0.0.11`) unless the user specifies otherwise.

### 2. Bump version in source

Edit `internal/version/version.go`:
```
const Version = "0.0.X"   # new version
```

### 3. Update version references in docs

**README.md** — exactly three occurrences (grep for the old version to confirm):
- `go install github.com/duber000/kukicha/cmd/kukicha@vOLD`
- `VERSION=vOLD`
- `**Version:** OLD — Ready for testing`

**CLAUDE.md** and **AGENTS.md** — one occurrence each near the top (must stay in sync):
- `Current version: **OLD**`

Replace all with the new version.

### 4. Regenerate and rebuild

```bash
make generate && make build
```

This regenerates `internal/semantic/stdlib_registry_gen.go`, all `stdlib/*/*.go`, all `stdlib/*/*_test.go`, then rebuilds the compiler with the updated files embedded.

### 5. Run tests

```bash
make test
```

All packages must pass before tagging. Do not proceed if any test fails.

### 6. Commit

Stage and commit in a single commit (per the contributing guide — do NOT split into multiple commits):

```bash
git add README.md internal/version/version.go stdlib/ internal/semantic/stdlib_registry_gen.go
git commit -m "chore: release vX.X.X

Bump version constant, update README install snippets, and regenerate
all stdlib .go and *_test.go files with the new version header."
```

### 7. Tag and confirm before pushing

Create the tag locally, then **ask the user to confirm** before pushing:

```bash
git tag vX.X.X
```

After user confirms:

```bash
git push origin main
git push origin vX.X.X
```

> Note: `git push --follow-tags` sometimes silently skips tags. Push the tag explicitly with `git push origin vX.X.X` and verify with `git ls-remote --tags origin | grep vX.X.X`.

## Checklist

- [ ] `internal/version/version.go` bumped
- [ ] `README.md` — all 3 version strings updated
- [ ] `CLAUDE.md` — version line updated
- [ ] `AGENTS.md` — version line updated (must match CLAUDE.md)
- [ ] `make generate && make build` succeeded
- [ ] `make test` — all packages pass
- [ ] Single commit with all generated + doc changes
- [ ] Tag created and pushed
- [ ] `git ls-remote --tags origin` confirms tag is present
