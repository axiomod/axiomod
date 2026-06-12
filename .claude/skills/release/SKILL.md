---
description: "Execute the release workflow: verify, tag, and prepare release. Use when the user says 'release', 'tag version', 'prepare release', 'cut release'."
---

# Release Workflow

## Inputs
- Version (required): semantic version (e.g., v1.5.0)

## Pre-release Checklist

1. **Verify build and tests**:
   ```bash
   make deps
   make build
   make build-cli
   make test
   make lint
   ```

2. **Verify architecture**:
   ```bash
   axiomod validator architecture
   ```

3. **Verify git state**:
   ```bash
   git status
   git log --oneline -5
   ```

## Release Steps

1. **Bump version in all files** using the bump script:
   ```bash
   ./scripts/bump-version.sh v<X.Y.Z>
   ```
   This updates:
   - `Makefile` (`VERSION :=`)
   - `.claude/CLAUDE.md` (version string)
   - `.claude/rules/15-ci-build.md` (version string)
   - `docs/release-notes/v<X.Y.Z>.md` (creates template if missing)
   - `docs/README.md` (adds release note link)

2. **Fill in release notes**: Edit `docs/release-notes/v<X.Y.Z>.md` with actual release content — replace the `[Title]` placeholder and describe what changed.

3. **Commit version bump**:
   ```bash
   git add Makefile .claude/CLAUDE.md .claude/rules/15-ci-build.md docs/release-notes/ docs/README.md
   git commit -m "chore: bump version to v<X.Y.Z>"
   ```

4. **Create tag** (the PreToolUse hook will automatically verify version sync):
   ```bash
   git tag v<X.Y.Z>
   ```
   > **Note**: A PreToolUse hook (`enforce-version-sync.sh`) will BLOCK this command if:
   > - Makefile `VERSION` does not match the tag
   > - `docs/release-notes/v<X.Y.Z>.md` does not exist
   >
   > If blocked, run `./scripts/bump-version.sh v<X.Y.Z>` and commit first.

5. **Push** (ask user for confirmation before executing):
   ```bash
   git push origin main
   git push origin v<X.Y.Z>
   ```

6. **Create GitHub release** (optional, ask user):
   ```bash
   gh release create v<X.Y.Z> --title "v<X.Y.Z>" --generate-notes
   ```

## Version Convention

- Semantic versioning: vMAJOR.MINOR.PATCH
- MAJOR: breaking API changes
- MINOR: new features, backward compatible
- PATCH: bug fixes
