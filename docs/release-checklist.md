# Axiomod Release Checklist

Follow this checklist to cut a new release of the Axiomod framework.
The current released version is tracked in `docs/release-notes/` (one file
per release) and must match `VERSION` in the `Makefile`.

## 1. Verify

- [ ] **Clean Dependencies**: Run `go mod tidy` and confirm no diff.
- [ ] **Verify Build**: Run `make build` and `make build-cli`.
- [ ] **Run Tests**: Run `go test -race ./...` — all green.
- [ ] **Lint Code**: Run `make lint` — no findings.
- [ ] **Architecture**: Run `axiomod validator architecture` — no violations.

## 2. Version

- [ ] Pick the next semantic version `vX.Y.Z`.
- [ ] Update `VERSION` in `Makefile` and `ARG VERSION` in `Dockerfile`.
- [ ] Update the version reference in `.claude/CLAUDE.md`.
- [ ] Write `docs/release-notes/vX.Y.Z.md` describing the changes.

## 3. Tag & Push

```bash
git add -A
git commit -m "Release vX.Y.Z"
git push origin main

git tag vX.Y.Z
git push origin vX.Y.Z
```

## 4. GitHub Release

1. Go to the GitHub repository.
2. Click **Releases** > **Draft a new release**.
3. Choose tag `vX.Y.Z`.
4. Title: `vX.Y.Z - <short summary>`.
5. Description: copy content from `docs/release-notes/vX.Y.Z.md`.
6. Attach binaries (if applicable).
