# Verifier Roadmap: Formal Proof Engineering for Kukicha

> Inspired by [Leanstral](https://mistral.ai/news/leanstral) — the insight that **small models + perfect verifiers > large models alone** applies directly to Kukicha's compiler-as-verifier architecture.

## Why This Matters

When AI agents are the primary code authors, every compile-time check is a proof obligation that all future generated code must satisfy. Kukicha's compiler already acts as a partial verifier (6 security checks, type system, Go compiler as second layer). This roadmap strengthens that verifier across three dimensions:

1. **AI-assisted code** — tighter generate-and-verify loops
2. **Agent runtime** — compiler as guardrail for agent-generated code
3. **Compiler correctness** — verifying the verifier itself

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

## Tier 1: Fuzz Testing (no new tooling)

**Goal:** Find crashes and panics in the compiler by throwing random inputs at it.

- [ ] **Fuzz the lexer** — Go's built-in `testing.F` (since Go 1.18). Feed random byte sequences to `lexer.New()` and verify it never panics, only returns errors.
  - File: `internal/lexer/lexer_fuzz_test.go`
  - Property: `Lex(input)` either produces valid tokens or returns errors, never panics

- [ ] **Fuzz the parser** — Feed random token sequences (or random source strings) to `parser.Parse()`.
  - File: `internal/parser/parser_fuzz_test.go`
  - Property: `Parse(input)` either produces a valid AST or returns parse errors, never panics

- [ ] **Fuzz the full pipeline** — Feed random `.kuki` source to the full lex→parse→semantic→codegen chain.
  - File: `internal/codegen/codegen_fuzz_test.go`
  - Property: pipeline either produces valid Go (verified by `go/parser`) or returns errors, never panics
  - Seed corpus: all existing test cases + `examples/*.kuki` + `stdlib/*.kuki`

---

## Tier 2: Property-Based Testing (no new tooling)

**Goal:** Verify algebraic properties of compiler transforms.

- [ ] **Roundtrip property** — `parse(format(parse(source))) == parse(source)` for the formatter.
  - Proves the formatter doesn't change program meaning
  - File: `internal/formatter/formatter_property_test.go`

- [ ] **Codegen structural soundness** — For any valid AST, `go/parser.ParseFile(codegen(ast))` succeeds.
  - Stronger than existing integration tests (random inputs, not hand-picked)
  - File: `internal/codegen/codegen_property_test.go`

- [ ] **Security check monotonicity** — Adding code to a program never removes a security error.
  - If `check(P)` reports error E, then `check(P + extra_code)` also reports E
  - File: `internal/semantic/security_property_test.go`

- [ ] **Return count consistency** — For all stdlib functions in `stdlib_registry_gen.go`, the declared return count matches the actual Go function signature in the generated `.go` file.
  - File: `internal/semantic/returncount_property_test.go`

---

## Tier 3: Structured Diagnostics for AI Agents

**Goal:** Make the compiler a better verifier for AI-in-the-loop workflows.

- [ ] **JSON error output** — `kukicha check --json` emits structured errors:
  ```json
  {
    "file": "app.kuki",
    "line": 12,
    "col": 5,
    "category": "security/sql-injection",
    "message": "string interpolation in SQL query argument",
    "suggestion": "use parameterized query: pg.Query(pool, \"SELECT * WHERE id = $1\", id)"
  }
  ```
  - Enables agents to parse errors programmatically and auto-fix

- [ ] **Fix suggestions for security errors** — Each security check category provides a concrete safe alternative.
  - SQL: "use $1 parameter placeholder"
  - XSS: "use http.Text() or template"
  - SSRF: "validate URL against allowlist with fetch.AllowedHosts()"
  - Shell: "use shell.Command() with explicit args"

- [ ] **Batch check mode** — `kukicha check --batch file1.kuki file2.kuki ...` for parallel verification of multiple candidate solutions.

---

## Tier 4: Specification Directives (new syntax)

**Goal:** Let Kukicha code express lightweight specifications that generate runtime checks.

- [ ] **`# kuki:ensures` directive** — Design by Contract for functions:
  ```kukicha
  # kuki:ensures result >= 0
  func Abs(x int) int
      if x < 0
          return -x
      return x
  ```
  Generates: `if !(result >= 0) { panic("ensures violated: result >= 0") }` in debug builds.

- [ ] **`# kuki:requires` directive** — Precondition checks:
  ```kukicha
  # kuki:requires len(items) > 0
  func First(items list of string) string
      return items[0]
  ```

- [ ] **`# kuki:invariant` on types** — Struct invariants checked after construction:
  ```kukicha
  # kuki:invariant self.min <= self.max
  type Range
      min int
      max int
  ```

- [ ] **Strip in release builds** — `kukicha build --release` omits all contract checks (zero runtime cost in production).

---

## Tier 5: Agent-Specific Security Checks

**Goal:** Extend the security check system for risks specific to AI agent code.

- [ ] **Unbounded loops** — Warn when a `for` loop has no obvious termination condition inside an HTTP handler or agent task.

- [ ] **Data exfiltration** — Flag when `fetch.Post`/`fetch.Get` is called with data that originated from `files.Read` or `os.ReadFile` inside an agent context.

- [ ] **Resource exhaustion** — Warn on unbounded `make(channel)`, unbounded `go` blocks, or allocation inside loops without size limits.

- [ ] **Privilege escalation** — Flag `shell.Run` or `shell.Command` calls that use variables derived from external input (HTTP params, env vars, file contents).

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

- [ ] **Explore Leanstral integration** — Use Leanstral (or similar) to assist with writing Lean 4 proofs about Kukicha's compiler properties. The meta-level: an AI proving that an AI-targeted compiler is correct.

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
