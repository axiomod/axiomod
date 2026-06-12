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
- [ ] Run `./scripts/bump-version.sh vX.Y.Z` (updates `Makefile`,
      `.claude/CLAUDE.md`, the CI/build rule, and seeds the release-note file).
- [ ] Update `ARG VERSION` in `Dockerfile`.
- [ ] Complete `docs/release-notes/vX.Y.Z.md` describing the changes.

## 3. Tag & Push

```bash
git add -A
git commit -m "Release vX.Y.Z"
git push origin main

git tag vX.Y.Z
git push origin vX.Y.Z
```

## 4. Automated Release (GoReleaser)

Pushing the tag triggers `.github/workflows/release.yml`, which runs the test
suite and then GoReleaser (`.goreleaser.yaml`):

- builds `axiomod` and `axiomod-server` for linux/darwin × amd64/arm64 with
  version ldflags,
- publishes the GitHub Release with archives and `checksums.txt`,
- generates a changelog from commit messages.

After the workflow finishes:

- [ ] Verify the release page lists all eight artifacts plus checksums.
- [ ] Verify `GOPROXY=direct go get github.com/axiomod/axiomod@vX.Y.Z`
      resolves from a scratch module.
- [ ] Paste the highlights from `docs/release-notes/vX.Y.Z.md` into the
      release description.
