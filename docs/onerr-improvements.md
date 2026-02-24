# OnErr UX Improvements (RFC Draft)

Status: Draft  
Audience: language design, parser/semantic/codegen maintainers, docs/tooling maintainers  
Last updated: 2026-02-24

## Summary

`onerr` is one of Kukicha's strongest features, but today it has avoidable UX friction:

- Too many forms to memorize early.
- Repetitive propagation boilerplate (`onerr return empty, error "{error}"`).
- Strict `{error}` naming rule causes frequent beginner mistakes (`{err}`).
- `onerr discard` is easy to misuse in production paths.

This RFC proposes incremental, backward-compatible changes:

1. Add a propagation shorthand: `onerr return`.
2. Add named error alias for block handlers: `onerr as e`.
3. Add style/lint rules to make risky handlers explicit (`discard`, silent defaults).
4. Improve diagnostics and docs around error variable naming.

## Goals

- Reduce repetition without removing explicit error behavior.
- Keep happy-path readability in pipe chains.
- Preserve compile-time safety guarantees.
- Minimize migration burden for existing Kukicha code.

## Non-Goals

- Replacing `onerr` with exceptions.
- Changing existing `onerr` forms or removing old syntax.
- Introducing runtime behavior changes for existing valid programs.

## Current Friction

### 1. Verbose propagate form

Common code today:

```kukicha
data := fetchData() onerr return empty, error "{error}"
```

This is precise but repetitive and visually heavy in long pipelines.

### 2. Hard-coded `{error}` name

Inside `onerr` handlers, the caught error variable is always `error`. This is consistent but unintuitive for users coming from `err`-style conventions.

### 3. Unsafe suppression patterns

`onerr discard` and some fallback defaults are valid and useful, but high-risk when generated automatically in critical paths.

## Proposal

## A) Propagation shorthand: `onerr return`

Allow:

```kukicha
data := fetchData() onerr return
```

Semantics:

- Equivalent to "propagate the original error with default zero-value return(s) for all non-error return positions".
- If the enclosing function has return signature `(T, error)`, emit equivalent of:
  - `return empty, error "{error}"`
- If the enclosing function has return signature `(error)`, emit equivalent of:
  - `return error "{error}"`
- If the enclosing function has no compatible error return, compiler error.

Benefits:

- Removes repetitive boilerplate.
- Makes "propagate upward" a first-class readable intent.

## B) Named error alias in block handlers: `onerr as <ident>`

Allow:

```kukicha
payload := fetchData() onerr as e
    print("fetch failed: {e}")
    return
```

Semantics:

- Introduces an interpolation alias for the caught error in this handler block.
- `{error}` remains valid for backward compatibility.
- `<ident>` applies only to this `onerr` block scope.

Compiler restrictions:

- Alias must be a valid identifier.
- Alias cannot shadow reserved keywords.
- Alias only available inside the handler block.

Benefits:

- Reduces beginner mistakes around `{err}`.
- Improves readability in larger handlers.

## C) Lint/strict policy for risky handlers

Add non-breaking warnings (promotable to errors in strict mode):

- `onerr discard` outside test files.
- Fallback literal defaults on security/IO/network/database operations without comment.
- `panic` inside library-like non-`main` entrypoints (warning only).

Suggested flags:

- `kukicha check` default: warnings.
- `kukicha check --strict-onerr`: treat warning set as errors.

This can be implemented in semantic analysis (same pass where security checks run), or as a dedicated lint pass invoked by `check`.

## D) Diagnostic upgrades

Improve common failures:

- If `{err}` appears in `onerr` context:
  - Current: compile error.
  - Proposed: compile error + fix hint:
    - "Use `{error}` or `onerr as e` and then `{e}`."
- If `onerr return` used in non-error-returning function:
  - Explain enclosing signature mismatch and suggest explicit fallback or `panic`.

## Grammar and AST Changes

## Grammar additions (conceptual)

```ebnf
onerr_clause
  = "onerr" (onerr_inline | onerr_block) ;

onerr_inline
  = "return"
  | existing_forms ;

onerr_block
  = ["as" identifier] NEWLINE INDENT statement* DEDENT ;
```

Notes:

- `onerr return` is a new inline form.
- `onerr as e` is only for block form.

## AST updates

Extend `OnErrClause` with:

- `ShorthandReturn bool`
- `Alias string` (empty if not set)

No breaking AST changes needed for existing forms.

## Semantic Analysis Changes

1. Validate `onerr return` compatibility with enclosing function return signature.
2. Populate `exprReturnCounts` behavior exactly as existing `onerr` paths.
3. For block handlers with alias:
   - Mark alias as valid interpolation variable in `onerr` scope.
   - Keep `{error}` valid.
4. Add warning diagnostics for risky `onerr` forms (feature-flagged strictness).

## Codegen Changes

1. `onerr return`:
   - Reuse existing generated error temp variable.
   - Emit appropriate `return ...` based on function signature metadata.
2. `onerr as e`:
   - Bind interpolation resolution for `{e}` to the same generated error temp var.
   - Preserve `{error}` mapping as synonym.

Implementation detail:

- Existing `currentOnErrVar` in generator already models dynamic onerr error variable binding and is a natural extension point.

## Backward Compatibility

- All existing valid code remains valid.
- `{error}` remains canonical and supported.
- New syntax is additive.
- Warnings are non-breaking unless strict mode is enabled.

## Rollout Plan

1. Implement parser + AST support for `onerr return` and `onerr as e`.
2. Implement semantic validation + improved diagnostics.
3. Implement codegen.
4. Add tests across lexer/parser/semantic/codegen.
5. Update docs:
   - `README.md` quick examples.
   - `docs/kukicha-quick-reference.md`.
   - Root `AGENTS.md` and `stdlib/AGENTS.md` onerr sections.
6. Ship warnings first, strict mode optional.

## Test Plan

Add tests for:

- Parse success: `onerr return`, `onerr as e`.
- Parse failure: malformed `onerr as`.
- Semantic success: shorthand return in `(T, error)` and `(error)` functions.
- Semantic failure: shorthand return where no error can be returned.
- Semantic interpolation: `{error}` and `{e}` both valid in alias blocks.
- Diagnostic quality: `{err}` includes direct fix suggestions.
- Codegen snapshots for representative inline and block forms.

## Open Questions

1. Should `onerr as e` also be allowed for inline forms for consistency?
2. Should `onerr return` preserve wrapped error context automatically, or always forward raw error?
3. Which operations should trigger "fallback default is risky" warnings by default?
4. Should strict-onerr be a dedicated flag or part of a broader strict profile?

## Recommendation

Adopt A + B immediately (high UX gain, low compatibility risk), ship C + D in the same release if warning defaults are conservative.
