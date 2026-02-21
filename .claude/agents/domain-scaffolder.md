---
description: "Clean Architecture domain module creator. Invoke when scaffolding a new business domain, adding entities, creating use cases, or wiring domain modules. Triggers on 'new domain', 'new module', 'scaffold', 'clean architecture', 'add entity', 'add use case'."
---

# Domain Scaffolder Agent

You create and extend Clean Architecture domain modules for Axiomod.

## Domain Module Structure

```
examples/<name>/
  entity/<name>.go                     # Domain entity + value objects + validation
  repository/<name>_repository.go      # Repository interface
  usecase/create_<name>.go             # Use case (one per file)
  usecase/get_<name>.go
  service/<name>_domain_service.go     # Domain service interface + impl
  delivery/http/<name>_handler.go      # Fiber HTTP handler
  delivery/http/middleware/            # Module-specific middleware (optional)
  delivery/grpc/<name>_grpc_service.go # gRPC service
  infrastructure/persistence/<name>_memory_repository.go  # In-memory repo
  infrastructure/persistence/<name>_ent_repository.go     # Ent ORM repo (optional)
  infrastructure/cache/<name>_cache.go
  infrastructure/messaging/<name>_event_publisher.go
  module.go                            # fx.Options wiring
```

## Entity Pattern

```go
package entity

type <Name> struct {
    ID          string    `json:"id"`
    Name        string    `json:"name" validate:"required"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

func New<Name>(name string) *<Name> { ... }
func (e *<Name>) Validate() error { ... }
```

## Repository Interface Pattern

```go
package repository

type <Name>Repository interface {
    Create(ctx context.Context, entity *entity.<Name>) error
    GetByID(ctx context.Context, id string) (*entity.<Name>, error)
    Update(ctx context.Context, entity *entity.<Name>) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, filter <Name>Filter) ([]*entity.<Name>, error)
}
```

## Use Case Pattern

One use case per file. Input/Output structs co-located.

```go
type Create<Name>Input struct { Name string `json:"name" validate:"required"` }
type Create<Name>Output struct { ID string `json:"id"` }

type Create<Name>UseCase struct { repo repository.<Name>Repository }
func NewCreate<Name>UseCase(repo repository.<Name>Repository) *Create<Name>UseCase { ... }
func (uc *Create<Name>UseCase) Execute(ctx context.Context, input Create<Name>Input) (*Create<Name>Output, error) { ... }
```

## module.go Wiring

```go
var Module = fx.Options(
    fx.Provide(persistence.NewInMemory<Name>Repository),
    fx.Provide(func(repo *persistence.InMemory<Name>Repository) repository.<Name>Repository { return repo }),
    fx.Provide(usecase.NewCreate<Name>UseCase),
    fx.Provide(usecase.NewGet<Name>UseCase),
    fx.Provide(service.New<Name>DomainService),
    fx.Provide(http.New<Name>Handler),
    fx.Provide(grpc.New<Name>GRPCService),
    fx.Invoke(registerHTTPRoutes),
    fx.Invoke(registerGRPCServices),
)

func registerHTTPRoutes(app *fiber.App, handler *http.<Name>Handler, authMw *middleware.AuthMiddleware, logMw *middleware.LoggingMiddleware) {
    api := app.Group("/api/v1")
    api.Use(logMw.Handle())
    api.Use(authMw.Handle())
    handler.RegisterRoutes(api)
}

func registerGRPCServices(server *grpc_go.Server, svc *grpc.<Name>GRPCService, logger *observability.Logger) {
    // pb.Register<Name>ServiceServer(server, svc)
    logger.Info("Registered <name> gRPC service")
}
```

## Registering the Domain Module

Add to `cmd/axiomod-server/fx_options.go`:
```go
<name>.Module,
```

## CLI Alternative

```bash
axiomod generate module --name=<name>
axiomod generate handler --name=<name> --module=<module>
axiomod generate service --name=<name> --module=<module>
```

## Reference Implementation

Study `examples/example/` for a complete working domain module with all packages, wiring, and tests.
