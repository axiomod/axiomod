# Architecture Rules

## Layer Structure (import direction: down only)

```
cmd/axiomod-server    -- entry point, fx assembly
cmd/axiomod           -- CLI (cobra commands)
plugins/              -- Plugin interface + registry + built-in plugins
platform/             -- server (Fiber HTTP + gRPC), observability
framework/            -- config, auth, middleware, errors, validation, health, DI
examples/             -- Clean Architecture domain modules
```

## Import Rules (from architecture-rules.json)

| Layer | Allowed Imports |
|---|---|
| `entity` | Nothing domain-level (stdlib + uuid only) |
| `repository` | `entity` only |
| `service` | `entity`, `repository` |
| `usecase` | `entity`, `repository`, `service` |
| `delivery/http` | `usecase`, `entity`, `middleware` |
| `delivery/grpc` | `usecase`, `entity` |
| `infrastructure/persistence` | `entity`, `repository` |
| `infrastructure/cache` | `entity` |
| `infrastructure/messaging` | `entity` |
| `platform/*` | `framework/*` |
| `plugins/*` | `platform/*`, `framework/*` |

## Forbidden

- **Cross-domain imports**: Domain module A must NEVER import from domain module B
- **Upward imports**: Lower layers must never import higher layers
- **Circular dependencies**: Not allowed anywhere

## Exceptions (exempt from import rules)

- `_test.go` files
- `mock_` prefixed files
- `testdata` directories
- `platform/ent/schema` and `platform/ent/migrate`

## Validation

Run `axiomod validator architecture` or `make lint` to verify compliance.

## Domain Module Structure (Clean Architecture, 8 packages)

```
examples/<name>/
  entity/                          # Domain entities, value objects, validation
  repository/                      # Repository interfaces, filter types, errors
  usecase/                         # One use case per file
  service/                         # Domain services
  delivery/http/                   # Fiber handlers
  delivery/grpc/                   # gRPC service implementations
  infrastructure/persistence/      # Repository implementations
  infrastructure/cache/            # Cache implementations
  infrastructure/messaging/        # Event publishers
  module.go                        # fx.Options wiring
```
