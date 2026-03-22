# Agent Tool Ergonomics — Design Reference

Captures the design decisions made in March 2026 for improving Kukicha's ergonomics when scripts are invoked as subprocess tools by AI agents.

---

## Context

Agents invoke Kukicha scripts as subprocess tools. The tool writes JSON to stdout; the agent parses it. Three friction points existed:

1. **Repeated boilerplate** — `scripts/changelog.kuki` and `examples/gh-semver-release/main.kuki` each defined an identical `printJSON` helper.
2. **Token waste** — both tools used `MarshalPretty` (2-space indented). Pretty JSON uses ~30–50% more tokens vs compact. Agents parse either identically.
3. **No skill discovery** — no stdlib mechanism to read `.agent/skills/` or `.claude/skills/` SKILL.md files at runtime, so an orchestrator written in Kukicha could not discover available tools dynamically.

---

## Changes

### 1. `stdlib/json` — `WriteOutput(v any) error`

Writes **compact** (not pretty) JSON to `os.Stdout`. Eliminates the `printJSON` helper from both call-site files.

```kukicha
result |> json.WriteOutput(_) onerr panic "{error}"
```

**Why compact not pretty**: `MarshalPretty` is for human reading. Compact is identical to agents and saves tokens at zero cost.

### 2. `stdlib/cli` — `IsJSON(args cli.Args) bool`

Shorthand for `cli.GetBool(args, "json")`. Makes the guard idiom one line:

```kukicha
if cli.IsJSON(args)
    result |> json.WriteOutput(_) onerr panic "{error}"
    return
```

### 3. `stdlib/skills` — New package

Reads SKILL.md manifests from `.agent/skills/` or `.claude/skills/` so an orchestrator written in Kukicha can discover available tools dynamically.

```kukicha
import "stdlib/skills"

available := skills.AgentSkills() onerr panic "{error}"
for s in available
    print("{s.Name}: {len(s.Content)} bytes")
```

`Discover(dir)` walks the directory, finds `SKILL.md` in each subdirectory, and returns a `Skill` (Name, Path, Content) for each. Returns an empty list — not an error — when the directory does not exist.

### 4. Call-site updates

- `scripts/changelog.kuki` — removed `printJSON`; replaced with `json.WriteOutput`; replaced `cli.GetBool(args, "json")` with `cli.IsJSON`; removed unused `import "fmt"`
- `examples/gh-semver-release/main.kuki` — same refactor pattern

### 5. Tutorial update

`docs/tutorials/agent-workflow-tutorial.md` — new "JSON output for agents" section covering the `WriteOutput` / `IsJSON` pattern and `stdlib/skills` for on-demand skill loading.

---

## Architecture Summary

| Concern | Where |
|---------|-------|
| Tool JSON output | `stdlib/json.WriteOutput` |
| `--json` flag check | `stdlib/cli.IsJSON` |
| Shell subprocess JSON | No change — `shell.Output(...) \|> parse.Json(...)` already works |
| A2A protocol | No change — `stdlib/a2a` is already a complete client |
| Skill reading at runtime | `stdlib/skills` |
| Token reduction | Compact output via `WriteOutput` + skills-on-demand pattern |

---

## Token Reduction Rationale

A typical `MarshalPretty` object with 5 fields and string values is ~120 characters. Compact JSON of the same object is ~80 characters — 33% smaller. For arrays of objects the difference compounds. Over hundreds of tool calls in a long agent session, this is significant.

Skills-on-demand compounds the savings: instead of embedding all SKILL.md files in the system prompt (often 2–10 KB total), the orchestrator loads only the skills it actually needs for the current task.

---

## Verification Checklist

1. `make test` — all existing tests pass
2. `make lint` — no new unused imports after call-site updates
3. `kukicha check scripts/changelog.kuki` — validates after refactor
4. `kukicha check examples/gh-semver-release/main.kuki` — validates after refactor
5. Manually run `kukicha run scripts/changelog.kuki --json` and verify compact JSON output (no indentation)
6. A test script calling `skills.AgentSkills()` in a directory without `.agent/skills/` returns an empty list with no error
