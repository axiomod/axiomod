# Contributing to Axiomod

Thank you for considering a contribution! This guide covers the workflow and
the quality gates every change must pass.

## Development Setup

```bash
git clone https://github.com/axiomod/axiomod.git
cd axiomod
make deps        # go mod tidy + download + pinned golangci-lint
make build       # bin/axiomod-server
make build-cli   # bin/axiomod
```

See [docs/developer-guide.md](docs/developer-guide.md) for the full guide.

## Quality Gates (run before every push)

```bash
make fmt             # gofmt
make lint            # golangci-lint (pinned version)
go vet ./...
go test -race ./...
make validate-arch   # architecture validator must report 0 violations
```

CI enforces all of the above plus a coverage threshold and a scaffold smoke
test (`axiomod init` output must compile).

## Pull Request Checklist

- [ ] Imports follow `architecture-rules.json` (run `make validate-arch`)
- [ ] Errors wrapped with `framework/errors` (no raw `fmt.Errorf` in
      framework code); entities use `DomainError`
- [ ] New fx modules registered in `cmd/axiomod-server/fx_options.go`
- [ ] Tests written (table-driven, `testify`), `go test -race ./...` green
- [ ] Docs updated in the same PR when behavior changes (`docs/`, `README.md`)
- [ ] A `CHANGELOG.md` entry under **Unreleased**

## Conventions

Coding style, naming, error handling, testing, and architecture rules are
documented in [`.claude/rules/`](.claude/rules/) and apply to human and AI
contributors alike. Highlights:

- Constructors are `NewXxx(deps...)`; middleware uses struct-with-`Handle()`
- Structured logging via `framework/observability` (never `fmt.Println` in
  library code)
- One use case per file with `Execute(ctx, input) (output, error)`
- Cross-domain imports are forbidden

## Commit Messages

Use conventional-commit style prefixes (`feat:`, `fix:`, `docs:`, `test:`,
`refactor:`, `chore:`); the release changelog is generated from them.

## Releases

Maintainers follow [docs/release-checklist.md](docs/release-checklist.md);
tags trigger the GoReleaser workflow.
