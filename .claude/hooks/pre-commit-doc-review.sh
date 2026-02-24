#!/usr/bin/env bash
# .claude/hooks/pre-commit-doc-review.sh
#
# PreToolUse hook: intercepts "git commit" commands and reminds the session
# to review documentation before committing.
#
# Returns a JSON deny decision so Claude updates docs first.
# NOTE: Do not use "set -euo pipefail" — a non-zero exit from any command
# causes Claude Code to treat the hook as errored and skip the deny decision.

# Read stdin (tool input JSON) and extract the command
INPUT=$(cat)
COMMAND=$(echo "$INPUT" | jq -r '.tool_input.command // ""' 2>/dev/null) || COMMAND=""

# Only intercept git commit commands — allow everything else
if [[ ! "$COMMAND" =~ ^[[:space:]]*git[[:space:]]+commit ]]; then
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

# Check if doc files are already staged (user/session already updated them)
STAGED_DOCS=$(git diff --cached --name-only -- \
    'CLAUDE.md' 'AGENTS.md' \
    'stdlib/CLAUDE.md' 'stdlib/AGENTS.md' \
    'internal/CLAUDE.md' 'internal/AGENTS.md' \
    '.claude/skills/**' '.agent/skills/**' 2>/dev/null) || true

if [ -n "$STAGED_DOCS" ]; then
    # Docs are already staged — allow the commit
    exit 0
fi

# Block the commit and ask the session to review docs
cat <<'DENY'
{"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"deny","permissionDecisionReason":"Documentation review required before committing. Staged code changes were detected but no documentation files are staged. Please review and update any affected CLAUDE.md/AGENTS.md files (root, stdlib/, internal/) and .claude/skills/ files, then stage them before committing. After updating, also copy each CLAUDE.md to the corresponding AGENTS.md in the same directory."}}
DENY
exit 0
