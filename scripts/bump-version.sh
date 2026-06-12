#!/usr/bin/env bash
# bump-version.sh — sync the project version across all files that embed it.
#
# Usage: ./scripts/bump-version.sh v<X.Y.Z>
#
# Updates:
#   - Makefile                      (VERSION := v<X.Y.Z>)
#   - .claude/CLAUDE.md             (version: v<X.Y.Z>)
#   - .claude/rules/15-ci-build.md  (framework/version.Version = v<X.Y.Z>)
#   - docs/release-notes/v<X.Y.Z>.md (creates template if missing)
#   - docs/README.md                (adds release note link if missing)
#
# The PreToolUse hook .claude/hooks/enforce-version-sync.sh blocks `git tag v*`
# until these files agree with the tag.

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"

NEW_VERSION="${1:-}"

if [[ ! "$NEW_VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+([-.][0-9A-Za-z.-]+)?$ ]]; then
  echo "Usage: $0 v<X.Y.Z>" >&2
  echo "Example: $0 v0.3.0" >&2
  exit 1
fi

OLD_VERSION=$(grep -E '^VERSION' "$REPO_ROOT/Makefile" | awk -F':=' '{print $2}' | tr -d ' ')

echo "Bumping version: $OLD_VERSION -> $NEW_VERSION"

# ── Makefile ──────────────────────────────────────────────────────────────────

sed -i.bak -E "s|^(VERSION[[:space:]]*:=[[:space:]]*)v[0-9][^ ]*|\1${NEW_VERSION}|" "$REPO_ROOT/Makefile"
rm -f "$REPO_ROOT/Makefile.bak"
echo "  updated Makefile"

# ── .claude/CLAUDE.md ─────────────────────────────────────────────────────────

sed -i.bak -E "s|(version: )v[0-9]+\.[0-9]+\.[0-9]+[^ ,\`]*|\1${NEW_VERSION}|" "$REPO_ROOT/.claude/CLAUDE.md"
rm -f "$REPO_ROOT/.claude/CLAUDE.md.bak"
echo "  updated .claude/CLAUDE.md"

# ── .claude/rules/15-ci-build.md ──────────────────────────────────────────────

sed -i.bak -E "s|(framework/version.Version\` = )v[0-9]+\.[0-9]+\.[0-9]+[^ ]*|\1${NEW_VERSION}|" "$REPO_ROOT/.claude/rules/15-ci-build.md"
rm -f "$REPO_ROOT/.claude/rules/15-ci-build.md.bak"
echo "  updated .claude/rules/15-ci-build.md"

# ── docs/release-notes/v<X.Y.Z>.md ────────────────────────────────────────────

RELEASE_NOTE="$REPO_ROOT/docs/release-notes/${NEW_VERSION}.md"

if [[ ! -f "$RELEASE_NOTE" ]]; then
  cat > "$RELEASE_NOTE" <<EOF
# ${NEW_VERSION} - [Title]

## 🌟 Key Highlights

* TODO: describe the most important changes in this release.

## 📦 What's New

* TODO: list new features, fixes, and improvements.

## ⬆️ Upgrade Notes

* TODO: note any breaking changes or migration steps (or remove this section).
EOF
  echo "  created docs/release-notes/${NEW_VERSION}.md (fill in the template)"
else
  echo "  docs/release-notes/${NEW_VERSION}.md already exists, skipping"
fi

# ── docs/README.md release note link ──────────────────────────────────────────

DOCS_README="$REPO_ROOT/docs/README.md"
LINK="- [${NEW_VERSION} - [Title]](./release-notes/${NEW_VERSION}.md)"

if ! grep -q "release-notes/${NEW_VERSION}.md" "$DOCS_README"; then
  awk -v link="$LINK" '
    /^## 📝 Release Notes/ { print; insert=1; next }
    insert && /^$/ { print; print link; insert=0; next }
    { print }
  ' "$DOCS_README" > "$DOCS_README.tmp" && mv "$DOCS_README.tmp" "$DOCS_README"
  echo "  added release note link to docs/README.md (update the [Title])"
else
  echo "  docs/README.md already links ${NEW_VERSION}, skipping"
fi

echo "Done. Review the changes, fill in the release notes, then commit."
