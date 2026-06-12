#!/usr/bin/env bash
# enforce-version-sync.sh — PreToolUse hook that blocks `git tag v*`
# if version files are out of sync.
#
# Hook protocol:
#   stdin  = JSON with { tool_name, tool_input: { command } }
#   exit 0 = ALLOW
#   exit 2 = BLOCK (stderr shown to user)

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"

# Read the hook input from stdin
INPUT=$(cat)

# Extract the command from tool_input.command
COMMAND=$(echo "$INPUT" | python3 -c "
import sys, json
data = json.load(sys.stdin)
print(data.get('tool_input', {}).get('command', ''))
" 2>/dev/null || echo "")

# Only intercept `git tag v*` commands
if ! echo "$COMMAND" | grep -qE '^\s*git\s+tag\s+v[0-9]'; then
  exit 0
fi

# Extract the tag version (first v* argument after `git tag`)
TAG_VERSION=$(echo "$COMMAND" | grep -oE 'v[0-9]+\.[0-9]+\.[0-9]+[^ ]*' | head -1)

if [[ -z "$TAG_VERSION" ]]; then
  exit 0
fi

# ── Check 1: Makefile VERSION matches the tag ────────────────────────────────

MAKEFILE_VERSION=$(grep -E '^VERSION' "$REPO_ROOT/Makefile" | awk -F':=' '{print $2}' | tr -d ' ')

if [[ "$MAKEFILE_VERSION" != "$TAG_VERSION" ]]; then
  echo "BLOCKED: Version mismatch — Makefile VERSION is $MAKEFILE_VERSION but tag is $TAG_VERSION" >&2
  echo "Run: ./scripts/bump-version.sh $TAG_VERSION" >&2
  exit 2
fi

# ── Check 2: Release notes file exists ───────────────────────────────────────

RELEASE_NOTE="$REPO_ROOT/docs/release-notes/${TAG_VERSION}.md"

if [[ ! -f "$RELEASE_NOTE" ]]; then
  echo "BLOCKED: Missing release notes at docs/release-notes/${TAG_VERSION}.md" >&2
  echo "Run: ./scripts/bump-version.sh $TAG_VERSION" >&2
  exit 2
fi

# All checks passed
exit 0
