# Verifier Roadmap: Formal Proof Engineering for Kukicha

> Inspired by [Leanstral](https://mistral.ai/news/leanstral) — Mistral's open-source Lean 4 proof agent.

## Key Insight from Leanstral

Leanstral is a **6B active parameter** model (highly sparse architecture) that outperforms Claude Opus 4.6, Sonnet 4.6, Qwen3.5 397B, Kimi-K2.5 1T, and GLM5 744B on formal proof engineering — despite being ~100x smaller. The secret: **parallel inference with Lean as a perfect verifier.**

The model generates many proof candidates simultaneously. Lean's kernel — a small, trusted piece of code — definitively accepts or rejects each one. No ambiguity, no "looks right." This generate-and-verify loop means a small model + perfect verifier > large model alone.

Crucially, Leanstral operates on **real repositories** (PRs to the Fermat's Last Theorem formalization project), not isolated competition problems. It can diagnose breaking changes across Lean versions, build test environments to reproduce failures, and propose fixes with correct rationale — the full engineering loop, not just theorem proving.

**The pattern:** Small, efficient code generation + a fast, trustworthy verifier in the loop + MCP for tool access = agent that punches far above its weight class.

## Why This Matters for Kukicha

When AI agents are the primary code authors, every compile-time check is a proof obligation that all future generated code must satisfy. Kukicha's compiler already acts as a partial verifier (6 security checks, type system, Go compiler as second layer). This roadmap strengthens that verifier across three dimensions:

1. **AI-assisted code** — tighter generate-and-verify loops (like Leanstral's parallel inference + Lean kernel)
2. **Agent runtime** — compiler as guardrail for agent-generated code (like Lean rejecting invalid proofs)
3. **Compiler correctness** — verifying the verifier itself (like Lean's trusted kernel being small and proven)

### The Analogy

| Leanstral | Kukicha |
|-----------|---------|
| Lean 4 kernel (perfect verifier) | `kukicha check` + Go compiler (partial verifier) |
| Proof candidates (parallel inference) | Code candidates (parallel `kukicha check`) |
| Lean's type theory (specifications) | Security checks + type system (partial specs) |
| `lean-lsp-mcp` (tool access) | Structured diagnostics + MCP (tool access) |
| FLT project PRs (real-world eval) | Real agent tasks on Kukicha codebases |

The gap: Lean is a *perfect* verifier (sound and complete for its logic). Kukicha's compiler is a *partial* verifier (catches classes of bugs but can't prove arbitrary correctness). This roadmap closes that gap incrementally.

---

## What We Already Have

| Verification Layer | What It Catches | Location |
|--------------------|-----------------|----------|
| 6 security checks | SQL injection, XSS, SSRF, path traversal, command injection, open redirect | `internal/semantic/semantic_security.go` |
| Type system (14 kinds) | Type mismatches, undefined types, struct field errors | `internal/semantic/semantic_types.go` |
| Return count inference | Incorrect `onerr` multi-value splits | `internal/semantic/semantic_declarations.go` |
| `kukicha check` | Fast syntax + semantic validation without compilation | `cmd/kukicha/` |
| Go compiler (2nd pass) | Anything Kukicha's analysis misses | Transpiled output |
| ~8,100 lines of tests | Regressions across lexer, parser, semantic, codegen, stdlib | `internal/*/` |

---

## Tier 1: Fuzz Testing (no new tooling) ✅

**Goal:** Find crashes and panics in the compiler by throwing random inputs at it.

- [x] **Fuzz the lexer** — Go's built-in `testing.F` (since Go 1.18). Feed random byte sequences to `lexer.New()` and verify it never panics, only returns errors.
  - File: `internal/lexer/lexer_fuzz_test.go`
  - Property: `Lex(input)` either produces valid tokens or returns errors, never panics

- [x] **Fuzz the parser** — Feed random token sequences (or random source strings) to `parser.Parse()`.
  - File: `internal/parser/parser_fuzz_test.go`
  - Property: `Parse(input)` either produces a valid AST or returns parse errors, never panics

- [x] **Fuzz the full pipeline** — Feed random `.kuki` source to the full lex→parse→semantic→codegen chain.
  - File: `internal/codegen/codegen_fuzz_test.go`
  - Property: pipeline either produces valid Go (verified by `go/parser`) or returns errors, never panics
  - Seed corpus: valid Kukicha programs covering functions, types, methods, pipes, onerr, etc.

---

## Tier 2: Property-Based Testing (no new tooling) ✅

**Goal:** Verify algebraic properties of compiler transforms.

- [x] **Formatter idempotency** — `format(format(source)) == format(source)` for the formatter.
  - Proves the formatter is stable (idempotent)
  - File: `internal/formatter/formatter_property_test.go`

- [x] **Codegen structural soundness** — For any valid AST, `go/parser.ParseFile(codegen(ast))` succeeds.
  - 20+ programs covering functions, types, methods, pipes, onerr, switch, lambdas, etc.
  - File: `internal/codegen/codegen_property_test.go`

- [x] **Security check monotonicity** — Adding code to a program never removes a security error.
  - All 6 security categories tested: SQL injection, XSS, SSRF, path traversal, command injection, open redirect
  - File: `internal/semantic/security_property_test.go`

- [x] **Return count consistency** — For all stdlib functions in both registries, `Count == len(Types)`.
  - Also checks: param names non-empty, security functions exist in registry, counts positive
  - File: `internal/semantic/returncount_property_test.go`

---

## Tier 3: Structured Diagnostics for AI Agents ✅

**Goal:** Make the compiler a better verifier for AI-in-the-loop workflows.

- [x] **JSON error output** — `kukicha check --json` emits structured errors:
  ```json
  {
    "file": "app.kuki",
    "line": 12,
    "col": 5,
    "severity": "error",
    "category": "security/sql-injection",
    "message": "SQL injection risk: ...",
    "suggestion": "use parameter placeholders ($1, $2, ...) instead of string interpolation"
  }
  ```
  - `internal/diagnostic/diagnostic.go` — Diagnostic struct implementing `error` interface
  - `internal/semantic/semantic.go` — `errorDiag`/`warnDiag` methods, `Diagnostics()` accessor
  - Enables agents to parse errors programmatically and auto-fix

- [x] **Fix suggestions for security errors** — Each security check category provides a concrete safe alternative.
  - `security/sql-injection`: "use parameter placeholders ($1, $2, ...) instead of string interpolation"
  - `security/xss`: "use http.SafeHTML to HTML-escape user-controlled content, or use http.Text() for plain text"
  - `security/ssrf`: "use fetch.SafeGet or add fetch.Transport(netguard.HTTPTransport(...))"
  - `security/path-traversal`: "use sandbox.* with a restricted root for user-controlled paths"
  - `security/command-injection`: "use shell.Output() with separate arguments for variable input"
  - `security/open-redirect`: "use http.SafeRedirect(w, r, url, allowedHosts...)"

- [x] **Batch check mode** — `kukicha check [--json] [--strict-onerr] file1.kuki file2.kuki ...`
  - Processes multiple files, collecting all diagnostics
  - JSON mode: outputs single JSON array with all diagnostics
  - Text mode: outputs errors grouped by file
  - The compiler becomes the "Lean kernel" in the agent loop

- [ ] **MCP server for kukicha check** — Expose the compiler as an MCP tool so any agent (including Leanstral-style small models) can call it natively.
  - Mirrors Leanstral's `lean-lsp-mcp` integration
  - Deferred: `--json` output can be piped through any MCP tool wrapper in the meantime

---

## Tier 4: Specification Directives (new syntax) ✅

**Goal:** Let Kukicha code express lightweight specifications that generate runtime checks.

- [x] **`# kuki:requires` directive** — Precondition checks:
  ```kukicha
  # kuki:requires "len(items) > 0"
  func First(items list of string) string
      return items[0]
  ```
  Generates: `if !(len(items) > 0) { panic("requires violated: len(items) > 0") }` at function entry.
  - File: `internal/codegen/codegen_contracts.go`
  - Semantic validation: `internal/semantic/semantic_contracts.go`
  - Kukicha keywords (`and`, `or`, `not`, `equals`) auto-translated to Go operators

- [x] **`# kuki:ensures` directive** — Design by Contract for functions:
  ```kukicha
  # kuki:ensures "result >= 0"
  func Abs(x int) int
      if x < 0
          return -x
      return x
  ```
  Generates: named return variables + `defer func() { if !(result >= 0) { panic("ensures violated: result >= 0") } }()` so postconditions are checked on all return paths.
  - File: `internal/codegen/codegen_contracts.go`
  - Semantic validation: rejects ensures on void functions (no return values)

- [x] **`# kuki:invariant` on types** — Struct invariants checked via generated `Validate()` method:
  ```kukicha
  # kuki:invariant "self.min <= self.max"
  type Range
      min int
      max int
  ```
  Generates: `func (r Range) Validate() { if !(r.min <= r.max) { panic("invariant violated: self.min <= self.max") } }`
  - File: `internal/codegen/codegen_contracts.go`
  - Semantic validation: rejects on type aliases, checks that `self.field` references exist on the struct

- [x] **Strip in release builds** — `kukicha build --release` omits all contract checks (zero runtime cost in production).
  - File: `cmd/kukicha/main.go` (new `--release` flag), `internal/codegen/codegen.go` (`releaseMode` field)

---

## Tier 5: Agent-Specific Security Checks ✅

**Goal:** Extend the security check system for risks specific to AI agent code.

- [x] **Unbounded loops** — Warn when `for true` has no break/return inside an HTTP handler.
  - Category: `agent/unbounded-loop`
  - File: `internal/semantic/semantic_security.go` (`checkUnboundedLoop`)

- [ ] **Data exfiltration** — Flag when `fetch.Post`/`fetch.Get` is called with data that originated from `files.Read` or `os.ReadFile` inside an agent context.
  - Deferred: requires taint tracking across variable assignments (non-trivial data flow analysis)

- [x] **Resource exhaustion** — Warn on goroutine spawning or channel creation inside loops in HTTP handlers.
  - Category: `agent/resource-exhaustion`
  - File: `internal/semantic/semantic_security.go` (`checkResourceExhaustion`)

- [x] **Privilege escalation** — Warn when `shell.Run` or `shell.Command` is called inside an HTTP handler (server privileges on user input).
  - Category: `agent/privilege-escalation`
  - File: `internal/semantic/semantic_security.go` (`checkPrivilegeEscalation`)

---

## Tier 6: Formal Verification (long-term)

**Goal:** Prove properties of the compiler itself using Lean 4 or similar.

- [ ] **Formalize type compatibility rules** — The `typesCompatible()` function in `semantic_types.go` has ~20 cases with subtle interactions. A Lean 4 specification could prove:
  - Reflexivity: `typesCompatible(T, T)` is always true
  - Symmetry where expected
  - No contradictions between cases

- [ ] **Prove security check soundness** — For a defined threat model, prove that the pattern-matching in `semantic_security.go` has no false negatives (sound) and characterize false positive rate.

- [ ] **Transpilation correctness for core constructs** — Formalize and prove semantic preservation for:
  - Pipe chains (`a |> b |> c` == `c(b(a))`)
  - `onerr` desugaring (error propagation preserves control flow)
  - String interpolation (generated `fmt.Sprintf` matches source semantics)

- [ ] **Explore Leanstral integration** — Use Leanstral (Apache 2.0, 6B active params, available via Mistral API) to assist with writing Lean 4 proofs about Kukicha's compiler properties. The meta-level: an AI proving that an AI-targeted compiler is correct.
  - Leanstral can operate on real repositories (not just isolated theorems) — feed it the Lean formalization of Kukicha's type rules and let it fill in proofs
  - Its sparse architecture makes it cost-efficient for iterative proof refinement

---

## Success Metrics

| Tier | Metric | Target |
|------|--------|--------|
| 1 | Fuzz test coverage | Lexer, parser, full pipeline; 0 panics after 10M iterations |
| 2 | Property violations found | Fix all violations; 0 remaining |
| 3 | Agent error-fix rate | Agents can auto-fix 80%+ of compiler errors using structured diagnostics |
| 4 | Contract adoption | All stdlib functions have `ensures`/`requires` where applicable |
| 5 | Agent security coverage | 4+ new check categories beyond current 6 |
| 6 | Formal proofs | Type compatibility + 1 security check category proven sound |
