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

3. **Check current version** in Makefile:
   ```bash
   grep "^VERSION" Makefile
   ```

4. **Verify git state**:
   ```bash
   git status
   git log --oneline -5
   ```

## Release Steps

1. **Update version** in Makefile `VERSION` field

2. **Commit version bump** (if changed):
   ```bash
   git add Makefile
   git commit -m "chore: bump version to v<X.Y.Z>"
   ```

3. **Suggest tag and push** (ask user for confirmation before executing):
   ```bash
   git tag v<X.Y.Z>
   git push origin main
   git push origin v<X.Y.Z>
   ```

4. **Create GitHub release** (optional, ask user):
   ```bash
   gh release create v<X.Y.Z> --title "v<X.Y.Z>" --generate-notes
   ```

## Version Convention

- Semantic versioning: vMAJOR.MINOR.PATCH
- MAJOR: breaking API changes
- MINOR: new features, backward compatible
- PATCH: bug fixes
