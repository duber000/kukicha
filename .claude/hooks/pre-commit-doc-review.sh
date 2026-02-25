#!/usr/bin/env bash
# .claude/hooks/pre-commit-doc-review.sh
#
# PreToolUse hook: intercepts "git commit" commands and requires the session
# to review documentation before committing.
#
# Blocks every code commit until the session explicitly confirms it has reviewed
# docs by prefixing the commit command with KUKICHA_DOCS_REVIEWED=1.
#
# Returns a JSON deny decision so Claude updates docs first.
# NOTE: Do not use "set -euo pipefail" — a non-zero exit from any command
# causes Claude Code to treat the hook as errored and skip the deny decision.

# Read stdin (tool input JSON) and extract the command
INPUT=$(cat)
COMMAND=$(echo "$INPUT" | jq -r '.tool_input.command // ""' 2>/dev/null) || COMMAND=""

# Only intercept git commit commands — allow everything else
if [[ ! "$COMMAND" =~ ^[[:space:]]*(KUKICHA_DOCS_REVIEWED=1[[:space:]]+)?git[[:space:]]+commit ]]; then
    exit 0
fi

# Check if the staged diff touches only doc files — if so, no review needed
STAGED_CODE=$(git diff --cached --name-only -- . \
    ':(exclude)*.md' \
    ':(exclude).claude/**' \
    ':(exclude).agent/**' 2>/dev/null) || true

if [ -z "$STAGED_CODE" ]; then
    exit 0
fi

# Allow if the session has explicitly confirmed doc review via the env var prefix
if [[ "$COMMAND" =~ ^[[:space:]]*KUKICHA_DOCS_REVIEWED=1 ]]; then
    exit 0
fi

# Block the commit and require explicit doc review confirmation
cat <<'DENY'
{"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"deny","permissionDecisionReason":"Documentation review required before committing. Review and update any affected CLAUDE.md/AGENTS.md files (root, stdlib/, internal/) and .claude/skills/ files. Copy each CLAUDE.md to its corresponding AGENTS.md in the same directory. Then commit with: KUKICHA_DOCS_REVIEWED=1 git commit ..."}}
DENY
exit 0
