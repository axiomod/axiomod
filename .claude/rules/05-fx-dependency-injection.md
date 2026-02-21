# fx Dependency Injection Patterns

## Module Declaration

Every package participating in DI exports a package-level `Module`:

```go
var Module = fx.Options(
    fx.Provide(NewLogger),
    fx.Provide(NewTracer),
    fx.Provide(NewMetrics),
    fx.Invoke(RegisterTracer),
)
```

Alternative named form for domain modules:

```go
var Module = fx.Module("domain_name",
    fx.Provide(http.NewDummyHandler),
    fx.Invoke(registerRoutes),
)
```

## Module Assembly

All modules assembled in `cmd/axiomod-server/fx_options.go`:

```go
func getModuleOptions() []fx.Option {
    return []fx.Option{
        observability.Module,
        middleware.Module,
        auth.Module,
        health.Module,
        grpc_pkg.Module,
        server.Module,
        plugins.Module,
        worker.Module,
        // Domain modules here
    }
}
```

**Every new module MUST be registered here.**

## Interface Binding

Bind concrete types to interfaces using anonymous provide functions:

```go
fx.Provide(func(repo *persistence.ExampleMemoryRepository) repository.ExampleRepository {
    return repo
}),
```

Or use the DI builder helpers:

```go
di.ProvideAs[repository.ExampleRepository](persistence.NewExampleMemoryRepository)
```

## Lifecycle Hooks

Register start/stop hooks for services with lifecycle management:

```go
func RegisterPlugins(lc fx.Lifecycle, registry *PluginRegistry) {
    lc.Append(fx.Hook{
        OnStart: func(ctx context.Context) error {
            return registry.StartAll()
        },
        OnStop: func(ctx context.Context) error {
            return registry.StopAll()
        },
    })
}
```

## Domain Module Wiring (module.go)

Standard pattern for a complete domain module:

```go
var Module = fx.Options(
    // Repositories
    fx.Provide(persistence.NewExampleMemoryRepository),
    fx.Provide(func(repo *persistence.ExampleMemoryRepository) repository.ExampleRepository {
        return repo
    }),
    // Use cases
    fx.Provide(usecase.NewCreateExampleUseCase),
    fx.Provide(usecase.NewGetExampleUseCase),
    // Domain services
    fx.Provide(service.NewExampleDomainService),
    // HTTP handlers
    fx.Provide(http.NewExampleHandler),
    // gRPC services
    fx.Provide(grpc.NewExampleGRPCService),
    // Route registration
    fx.Invoke(registerHTTPRoutes),
    fx.Invoke(registerGRPCServices),
)
```

## Rules

1. One `var Module` per package that provides DI components
2. Use `fx.Provide` for constructors, `fx.Invoke` for side-effect registration
3. Always register new modules in `fx_options.go`
4. Use `fx.Lifecycle` for start/stop hooks, not `init()` functions
5. Constructor signature determines what fx injects -- accept interfaces, return concrete types
