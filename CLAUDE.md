# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

**Axiomod** is an enterprise Go macroservice framework. Module path `github.com/axiomod/axiomod`, Go 1.24.2, version v1.4.0. Dual-protocol (HTTP via Fiber + gRPC), `go.uber.org/fx` dependency injection, plugin architecture, Clean Architecture for domain modules.

Two binaries:
- `bin/axiomod-server` — runtime server (entry: `cmd/axiomod-server/main.go`)
- `bin/axiomod` — CLI for scaffolding, migrations, plugin management, RBAC policies (entry: `cmd/axiomod/main.go`)

## Common Commands

```bash
make build          # Build server -> bin/axiomod-server (with version ldflags)
make build-cli      # Build CLI    -> bin/axiomod
make all            # Build both
make test           # go test -v ./...
make lint           # golangci-lint run ./...   (no .golangci.yml; uses defaults)
make fmt            # go fmt ./...
make deps           # go mod tidy + download + install golangci-lint
make docker         # docker build -t axiomod/server:latest .
make clean
```

Single-package test: `go test -v ./framework/errors/...`
Race detector (CI uses this; Makefile does not): `go test -race ./...`
Single test by name: `go test -v -run TestName ./path/...`
Benchmarks: `go test -bench=. ./framework/errors/`
Architecture validation: `axiomod validator architecture` (or `./bin/axiomod validator architecture --config=architecture-rules.json`)
Run server locally: `go run cmd/axiomod-server/main.go [-config=path]`

CLI subcommands (see `cmd/axiomod/cmd/`): `init`, `generate {module,handler,service}`, `migrate {up,down,create,force,version}`, `plugin {install,list,remove}`, `policy {add,list,remove}`, `validator architecture`.

## High-Level Architecture

The codebase is split into strict layers with one-way (downward) imports. Understanding the layer boundaries is essential — they're enforced by `architecture-rules.json` and validated by `axiomod validator architecture`.

```
cmd/axiomod-server/   entry point + fx assembly (fx_options.go wires every module)
cmd/axiomod/          cobra-based CLI
plugins/              Plugin interface + registry + built-in plugins (mysql, postgres, jwt, keycloak, casdoor, casbin, audit, ldap, saml, multitenancy, elk)
platform/             server (Fiber HTTP + gRPC), observability (zap/OTel/Prometheus)
framework/            config, auth, middleware, errors, validation, health, DI, resilience, kafka, worker, http client, etc.
examples/             reference Clean Architecture domain modules (the 8-package layout)
configs/              service_default.yaml + prometheus.yml + env files
```

Allowed import direction: `examples/<domain>` → `platform/*` → `framework/*`. `plugins/*` may import `platform/*` and `framework/*`. **Cross-domain imports between sibling `examples/<domain>` packages are forbidden.** See `architecture-rules.json` for the full matrix; exceptions are `_test.go`, `mock_*`, `testdata`, `platform/ent/{schema,migrate}`.

### fx Wiring is the Backbone

Every package that participates in DI exports a package-level `var Module = fx.Options(...)` (or `fx.Module("name", ...)`). The full app is assembled in `cmd/axiomod-server/fx_options.go::getModuleOptions()`. **A new module is invisible until added there.** Domain modules typically wire `fx.Provide` for repos/usecases/handlers and `fx.Invoke` for `registerHTTPRoutes` / `registerGRPCServices`. Lifecycle hooks (`fx.Lifecycle.Append`) replace `init()` for start/stop logic — the plugin registry uses this to call `Start()`/`Stop()` on all enabled plugins.

### Domain Module Shape (Clean Architecture, 8 packages)

`examples/example/` is the canonical reference. Each domain follows:

```
entity/                          # Pure domain types, validation, DomainError values
repository/                      # Repository interfaces, filter types, repo errors
service/                         # Cross-entity domain logic (entity + repository)
usecase/                         # One use case per file: Execute(ctx, Input) (*Output, error)
delivery/http/                   # Fiber handlers: NewXxxHandler + RegisterRoutes(fiber.Router)
delivery/grpc/                   # gRPC service impls (embed UnimplementedXxxServer)
infrastructure/persistence/      # Repository implementations (memory, Ent)
infrastructure/cache/            # Cache implementations
infrastructure/messaging/        # Event publishers (topic: "<entity>.<eventType>")
module.go                        # fx.Options wiring everything above
```

Handlers and gRPC methods contain **no business logic** — they parse, delegate to a use case, and format the response. Use cases orchestrate entities/repositories. Memory repos use `sync.RWMutex` and deep-clone entities to prevent shared state.

### Plugin System

`plugins/plugin.go` defines the 4-method `Plugin` interface (`Name`, `Initialize`, `Start`, `Stop`). Plugins are registered in `cmd/axiomod-server/register_plugins.go::RegisterNewPlugins`, configured via `configs/service_default.yaml` under `plugins.enabled` / `plugins.settings`, and lifecycle-managed by the registry through fx hooks. Plugins should call `health.RegisterCheck` in `Initialize`.

### Config

Single source of truth: `framework/config/types.go::Config{App, Observability, Database, HTTP, GRPC, Auth, Casbin, Plugins}`. Loaded by Viper from `configs/service_default.yaml`. Env override prefix is `APP_` with dots replaced by underscores (e.g. `APP_HTTP_PORT`). **YAML keys are camelCase with lowercased acronyms** (`issuerUrl`, `clientId`, `jwksCacheTtl`); **Go struct fields are PascalCase with capitalized acronyms** (`IssuerURL`, `ClientID`, `JWKSCacheTTL`). No `mapstructure`/`yaml` tags on config structs — Viper matches case-insensitively.

### Errors

Use `framework/errors` everywhere in framework/platform/application code (never raw `fmt.Errorf` except for low-level infrastructure wrapping). Constructors: `errors.New/Wrap/WithCode/WithMetadata` plus shorthands like `NewNotFound`, `NewInvalidInput`, `NewUnauthorized`, `NewForbidden`, `NewConflict`, `NewInternal`. `errors.ToHTTPCode(err)` and `errors.ToGRPCCode(err)` map codes to protocol statuses. Entity-level violations use `entity.NewDomainError("<domain>.<snake_case>", "msg")`.

### Observability

`platform/observability` exports `Logger` (zap wrapper), `Tracer` (OTel wrapper), `Metrics` (Prometheus registry with pre-defined vectors: `HTTPRequestsTotal`, `HTTPRequestDuration`, `GRPCRequestsTotal`, `GRPCRequestDuration`, `DBQueryDuration`). Endpoints `/live`, `/ready`, `/health`, `/metrics` are wired by the platform server. **Never use `fmt.Println`/`log.Println`/`zap.L()`** — always inject `*observability.Logger` and use structured fields with snake_case keys.

### Middleware

Struct-with-`Handle()` pattern: `type X struct{...}; func (m *X) Handle() fiber.Handler`. Server-level middleware (recover, cors, compress, logger, metrics, tracing) is applied in `platform/server/server.go`. Domain-level (auth, RBAC, logging) is applied per route group in the domain's `module.go`.

### Resilience

`framework/circuitbreaker`, `framework/resilience` (combines CB + retry + timeout + fallback), `framework/client` (HTTP client with built-in resilience). Wrap **all external calls** with these.

## Key Conventions (project-wide)

- Constructors: `NewXxx(deps...) *Xxx` or `(*Xxx, error)`
- Validation: `validator/v10` struct tags on use case `Input` types; framework wrapper at `framework/validation.Validator`
- JSON tags: **snake_case** for framework/API and gRPC types; **camelCase** for use case `Input`/`Output`
- Logging field keys, Prometheus labels, Fiber `c.Locals` keys: **snake_case**
- Plugin names in config: lowercase single words (`postgres`, `jwt`, `keycloak`)
- Imports grouped stdlib / internal (`github.com/axiomod/axiomod/...`) / third-party with blank lines between
- Tests: `testify/assert` + `require`, table-driven `t.Run`, manual mocks (no gomock/mockery), fxtest for DI integration
- Coverage target: >80% for core framework packages

## Detailed Rules

The `.claude/` directory contains additional guidance that takes precedence over this overview when there is overlap:

- `.claude/CLAUDE.md` — full project rules summary
- `.claude/rules/01-coding-style.md` through `15-ci-build.md` — per-topic deep dives (architecture, errors, testing, fx DI, HTTP/gRPC delivery, plugins, domain patterns, middleware, config, observability, resilience, database, CI)
- `.claude/agents/` — specialized subagents (arch-guard, domain-scaffolder, plugin-builder, security-scanner, etc.)
- `.claude/skills/` — invocable skills: `gen-module`, `add-endpoint`, `new-plugin`, `validate-arch`, `release`

Read the relevant `.claude/rules/*.md` file before making non-trivial changes in that area.

## Pre-Submit Checklist

- [ ] Imports respect `architecture-rules.json` (run `axiomod validator architecture`)
- [ ] Errors use `framework/errors` (not raw `fmt.Errorf`) in framework/app code
- [ ] New fx modules registered in `cmd/axiomod-server/fx_options.go`
- [ ] Tests added (table-driven, `testify`)
- [ ] `make fmt && make lint && make test` all pass
