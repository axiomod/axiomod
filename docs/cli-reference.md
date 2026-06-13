# Axiomod CLI Reference

The `axiomod` CLI is the primary tool for developing, managing, and deploying
applications built with the Axiomod framework. This reference matches the
binary's `--help` output; when in doubt, the binary is the source of truth.

## Installation

```bash
make build-cli
# The binary will be available at ./bin/axiomod
```

## Global Flags

| Flag | Description |
|---|---|
| `--config` | Path to the CLI configuration file (default `$HOME/.axiomod.yaml`) |
| `--help`   | Show help for any command |

## Project Scaffolding

### `init`

```bash
axiomod init my-service
axiomod init my-service --dev --framework-path /path/to/axiomod
```

Creates a new project: `cmd/<name>/main.go` (full fx assembly),
`configs/service_default.yaml`, `Dockerfile`, `Makefile`, `README.md`,
`.gitignore`, `LICENSE`, and the `internal/`, `tests/`, `migrations/`
directory skeleton. The generated `go.mod` pins the framework version this
CLI was built from, so `go mod tidy && go build ./...` works immediately.

| Flag | Description |
|---|---|
| `--dev` | Develop against a local framework checkout via a `replace` directive |
| `--framework-path` | Path to the local checkout (with `--dev`; default: discovered by walking up from the current directory) |

## Code Generation (`generate`)

All generated Go files are gofmt-clean.

```bash
axiomod generate module --name=order            # 8-package Clean Architecture module under examples/<name>/
axiomod generate service --name=payment --module=billing
axiomod generate handler --name=order --module=order
```

| Command | Flags |
|---|---|
| `generate module` | `--name` (required) |
| `generate service` | `--name` (required), `--module` (optional) |
| `generate handler` | `--name` (required), `--module` (optional, defaults to handler name) |

## Database Migrations (`migrate`)

Backed by golang-migrate; supports **postgres** and **mysql** (per
`database.driver`). Migration files live in `./migrations`.

```bash
axiomod migrate create add_users_table   # timestamped .up.sql / .down.sql pair
axiomod migrate up                       # apply pending migrations (creates the DB if missing)
axiomod migrate down [N]                 # roll back N steps (default 1)
axiomod migrate force <version>          # force-set version (dirty-state recovery)
axiomod migrate version                  # print current version + dirty flag
```

## Policy Management (`policy`)

Real Casbin operations against the model/policy configured under `casbin:`
(defaults ship at `configs/rbac_model.conf` / `configs/rbac_policy.csv`).
Arguments are positional:

```bash
axiomod policy list                          # print g (roles) and p (permissions) rules
axiomod policy add p <sub> <obj> <act>       # add a permission
axiomod policy add g <user> <role>           # assign a role
axiomod policy remove p <sub> <obj> <act>    # remove a permission
axiomod policy remove g <user> <role>        # remove a role assignment
```

## Configuration (`config`)

```bash
axiomod config validate     # YAML-validate configs/, config/, framework/config (incl. configs/env/)
axiomod config diff dev prod # compare configs/env/<env>.yaml overlays key by key
```

`config validate` exits non-zero when any file fails to parse; `config diff`
exits non-zero when an environment file cannot be found.

## Development Workflow

```bash
axiomod build               # builds cmd/<module> (or cmd/axiomod-server) to bin/
axiomod test                # go test -v -cover ./...
axiomod test ./pkg/...      # specific target
axiomod test --unit         # only ./tests/unit/...
axiomod test --integration  # only ./tests/integration/... with RUN_INTEGRATION_TESTS=true
axiomod lint                # golangci-lint (auto-installs if missing)
axiomod fmt                 # gofmt -w .
```

## DevOps

### `dockerize`

```bash
axiomod dockerize                       # image tag <module-name>:latest
axiomod dockerize --tag=myapp:v1.0.0
```

Generates a multi-stage Dockerfile (Go 1.25 builder, non-root Alpine runtime,
healthcheck, `configs/` baked in) and runs `docker build`.

### `deploy`

```bash
axiomod deploy dev                       # DRY RUN (default): builds the image, prints [SIMULATED] steps
axiomod deploy prod --dry-run=false \
    --registry registry.example.com/team \
    --manifests deploy/kubernetes        # real docker push + kubectl apply
```

| Flag | Description |
|---|---|
| `--dry-run` | Default `true`: build only, label remaining steps `[SIMULATED]` |
| `--registry` | Registry prefix for the push (required with `--dry-run=false`) |
| `--manifests` | Manifest path for `kubectl apply` (optional) |

### `status` / `healthcheck`

```bash
axiomod status                           # GET /ready, prints per-component health, exit 1 if unhealthy
axiomod status --url http://host:8080
axiomod healthcheck                      # simple GET /health probe
```

## Plugins (`plugin`)

```bash
axiomod plugin list                  # scans plugins/ for standalone plugins + lists built-ins
axiomod plugin install <git-url>     # clones into plugins/<name>; registration is manual
axiomod plugin remove <name>         # removes plugins/<name>; unregistration is manual
```

After install/remove, wire the plugin in `RegisterNewPlugins`
(`cmd/axiomod-server/register_plugins.go`) and toggle it under
`plugins.enabled` in `configs/service_default.yaml`.

## Validators (`validator`)

| Command | What it does |
|---|---|
| `validator architecture [--config=rules.json]` | AST-based import-rule validation against `architecture-rules.json` (also `make validate-arch`; enforced in CI) |
| `validator domain` | Cross-domain boundary check |
| `validator naming [--fix] [--json] [--sql DIR] [--api DIR]` | Go/API/SQL naming conventions |
| `validator security` | gosec (auto-installs) |
| `validator static-check` | staticcheck |
| `validator static-analysis` | go vet + staticcheck + gosec |
| `validator check-api-spec --spec FILE` | OpenAPI 3.x validation (kin-openapi); invalid specs exit non-zero |
| `validator check-docs [--since=REF]` | warns when code changed without documentation updates (git-diff heuristic) |
| `validator standards-check` | runs the full validator suite |

## Misc

```bash
axiomod version       # version, commit, build date, Go version, platform
axiomod interactive   # REPL: run any subcommand from an axiomod> prompt
axiomod completion    # shell completion scripts (bash/zsh/fish/powershell)
```
