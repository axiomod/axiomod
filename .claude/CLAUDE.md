# Axiomod Framework

Enterprise Go macroservice framework. Dual-protocol (HTTP/Fiber + gRPC), fx DI, plugin architecture.

## Quick Commands

```
make build          # Build server -> bin/axiomod-server
make build-cli      # Build CLI -> bin/axiomod
make test           # go test -v ./...
make lint           # golangci-lint run ./...
make fmt            # go fmt ./...
make deps           # go mod tidy + download + install golangci-lint
make docker         # docker build -t axiomod/server:latest .
```

### CLI (axiomod)

```
axiomod generate module --name=<name>                    # Scaffold 8-package domain
axiomod generate handler --name=<name> --module=<mod>    # HTTP handler + service + entity
axiomod generate service --name=<name> --module=<mod>    # Domain service + repository
axiomod validator architecture [--config=rules.json]     # Validate import rules
axiomod init <project-name>                              # Scaffold new project
axiomod migrate up|down|create|force|version             # Database migrations
axiomod plugin install|list|remove                       # Plugin management
axiomod policy add|list|remove                           # Casbin RBAC policies
```

## Architecture

### Layer Structure (import direction: down only)

```
cmd/axiomod-server    -- entry point, fx assembly
cmd/axiomod           -- CLI (cobra commands)
plugins/              -- Plugin interface + registry + built-in plugins
platform/             -- server (Fiber HTTP + gRPC), observability (zap, OTel, Prometheus)
framework/            -- config, auth, middleware, errors, validation, health, DI
examples/             -- Clean Architecture domain modules (reference implementations)
```

### Import Rules (architecture-rules.json)

- `entity` imports nothing domain-level
- `repository` -> entity only
- `usecase` -> entity, repository, service
- `service` -> entity, repository
- `delivery/http` -> usecase, entity, middleware
- `delivery/grpc` -> usecase, entity
- `infrastructure/*` -> entity, repository
- `platform/*` may import `framework/*`
- `plugins/*` may import `platform/*` and `framework/*`
- Cross-domain imports are **forbidden**
- Exceptions: `_test.go`, `mock_`, `testdata`, `platform/ent/schema`, `platform/ent/migrate`

## fx Module Pattern

Every module exports `var Module = fx.Options(...)`:

```go
var Module = fx.Options(
    fx.Provide(NewFoo),
    fx.Invoke(RegisterHooks),
)
```

Assembly in `cmd/axiomod-server/fx_options.go`. Current modules: observability, middleware, auth, health, grpc, server, plugins, worker. Domain modules wire via `fx.Invoke(registerHTTPRoutes)` + `fx.Invoke(registerGRPCServices)`.

## Plugin System

Interface at `plugins/plugin.go`:

```go
type Plugin interface {
    Name() string
    Initialize(config map[string]interface{}, logger *observability.Logger,
               metrics *observability.Metrics, cfg *config.Config, health *health.Health) error
    Start() error
    Stop() error
}
```

Register via `PluginRegistry.Register(&YourPlugin{})`. Built-in: mysql, postgresql, jwt, keycloak, casdoor, casbin. Enable/disable via `configs/service_default.yaml` under `plugins.enabled`.

## Config System

- Viper-based, reads `configs/service_default.yaml`
- Env prefix: `APP_` (e.g., `APP_HTTP_PORT=8080`)
- Struct: `framework/config/types.go` -> `Config{App, Observability, Database, HTTP, GRPC, Auth, Casbin, Plugins}`

## Error Handling

Use `framework/errors`: `errors.New()`, `errors.Wrap()`, `errors.WithCode()`, `errors.NewNotFound()`, etc. Maps to HTTP via `errors.ToHTTPCode()` and gRPC via `errors.ToGRPCCode()`. Codes: NOT_FOUND, INVALID_INPUT, UNAUTHORIZED, FORBIDDEN, INTERNAL_ERROR, TIMEOUT, UNAVAILABLE, CONFLICT, ALREADY_EXISTS.

## Domain Module Structure (Clean Architecture)

```
examples/<name>/
  entity/              # Domain entities, value objects, validation
  repository/          # Repository interfaces
  usecase/             # One use case per file: Execute(ctx, input) (output, error)
  service/             # Domain services
  delivery/http/       # Fiber handlers: NewXxxHandler + RegisterRoutes(fiber.Router)
  delivery/grpc/       # gRPC service implementations
  infrastructure/persistence/  # Repository implementations (memory, Ent)
  infrastructure/cache/        # Cache implementations
  infrastructure/messaging/    # Event publishers
  module.go            # fx.Options wiring all packages
```

## Observability

- Logger: `observability.Logger` wraps `zap.Logger` -- `logger.Info("msg", zap.String("k", "v"))`
- Tracer: `observability.Tracer` wraps OTel -- `tracer.Tracer.Start(ctx, "span")`
- Metrics: `observability.Metrics` wraps Prometheus -- `metrics.HTTPRequestsTotal.WithLabelValues(...).Inc()`
- Endpoints: `/live`, `/ready`, `/health`, `/metrics`

## Testing Conventions

- `testify/assert` + `testify/require`, table-driven with `t.Run()`
- fx integration: `fxtest.New(t, fx.Options(...))`
- Race detector: `go test -race ./...`
- Test files: `*_test.go` next to source
- Single package: `go test -v ./framework/errors/...`

## Key Conventions

- Go 1.24.2, module: `github.com/axiomod/axiomod`, version: v1.4.0
- Constructors: `NewXxx(deps...) *Xxx` or `NewXxx(deps...) (*Xxx, error)`
- Validation: `validator/v10` struct tags + `framework/validation.Validator`
- Middleware: struct with `Handle() fiber.Handler` method
- Health: `health.RegisterCheck(name, func() error)`
- No `fmt.Println` -- use `observability.Logger`
- No hardcoded config -- use `*config.Config`
- No business logic in handlers -- delegate to use cases

## Post-Implementation Checklist

- [ ] Imports follow `architecture-rules.json` direction
- [ ] Errors wrapped with `framework/errors` (not raw `fmt.Errorf`)
- [ ] New modules registered in `fx_options.go`
- [ ] Tests written (table-driven, `testify`)
- [ ] `make lint` passes
- [ ] `make test` passes
