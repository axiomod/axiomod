# Coding Style & Naming Conventions

## Go Type Naming

| Kind | Convention | Examples |
|---|---|---|
| Structs | PascalCase | `ExampleHandler`, `PluginRegistry`, `CircuitBreaker` |
| Interfaces | PascalCase, behavior-describing | `Plugin`, `ExampleRepository`, `Publisher`, `Cache`, `Provider` |
| Constructors | `New<Type>` prefix | `NewJWTService`, `NewExampleCache` |
| Default funcs | `Default<Name>` / `Default<Name>Options` | `DefaultConfig()`, `DefaultRetryOptions()` |
| Methods (exported) | PascalCase | `Handle()`, `RegisterRoutes()`, `Execute()` |
| Methods (unexported) | camelCase | `captureStack()`, `cloneExample()` |
| Config structs | `<Domain>Config` suffix | `AppConfig`, `HTTPConfig`, `DatabaseConfig` |
| Acronyms in Go | Fully capitalized | `HTTPConfig`, `GRPCConfig`, `IssuerURL`, `ClientID`, `JWKSCacheTTL`, `SSLMode` |

## File Naming

- **Source files**: snake_case (`example_handler.go`, `example_domain_service.go`)
- **Test files**: `*_test.go` co-located with source
- **Module wiring**: `module.go` per package

## Variable Naming

| Kind | Convention | Examples |
|---|---|---|
| Error sentinels | `Err<Description>` | `ErrNotFound`, `ErrInvalidToken`, `ErrJobNotFound` |
| Error code constants | `Code<Name>` PascalCase, `SCREAMING_SNAKE_CASE` values | `CodeNotFound = "NOT_FOUND"` |
| Health status | `Status<Value>` | `StatusUp = "UP"`, `StatusDown = "DOWN"` |
| Circuit breaker states | `State<Value>` iota | `StateClosed`, `StateOpen`, `StateHalfOpen` |
| Exporter types | `ExporterType<Value>` | `ExporterTypeJaeger = "jaeger"` |
| Event constants | `<Entity><Action>Event` | `ExampleCreatedEvent = "example.created"` |
| Domain error codes | `<domain>.<snake_case>` string values | `"example.empty_name"`, `"repository.example_not_found"` |
| fx module var | Always `var Module` | `var Module = fx.Options(...)` |
| Version vars | Exported PascalCase | `Version`, `GitCommit`, `BuildDate` |
| CLI command vars | Unexported camelCase | `rootCmd`, `generateModuleCmd`, `migrateCmd` |
| Template constants | Unexported camelCase | `entityTemplate`, `repositoryTemplate` |
| Plugin names | Lowercase single words | `"postgres"`, `"jwt"`, `"keycloak"`, `"multitenancy"` |

## JSON Tag Naming

| Context | Convention | Example |
|---|---|---|
| Framework/API structs | **snake_case** | `json:"user_id"`, `json:"created_at"`, `json:"git_commit"` |
| Use case Input/Output | **camelCase** | `json:"valueType"`, `json:"createdAt"` |
| gRPC/Protobuf types | **snake_case** | `json:"value_type,omitempty"` |

## Structured Logging Keys

**Convention: snake_case** for all zap field keys:

```go
zap.String("method", method)
zap.String("user_agent", ua)
zap.String("query_type", qt)
zap.Int("status", status)
```

**Never** use camelCase or kebab-case for log field keys.

## Fiber Context Locals Keys

**Convention: snake_case**:

```go
c.Locals("user_id", ...)
c.Locals("username", ...)
c.Locals("roles", ...)
```

## Prometheus Metric Naming

Follow Prometheus conventions:
- Metric names: **snake_case** with `_total` (counters) or `_seconds` (histograms) suffix
- Label names: **snake_case**
- Go fields: PascalCase

```
http_requests_total           -> HTTPRequestsTotal
http_request_duration_seconds -> HTTPRequestDuration
grpc_requests_total           -> GRPCRequestsTotal
db_query_duration_seconds     -> DBQueryDuration
```

Labels: `method`, `path`, `status`, `service`, `query_type`

## Import Organization

Three-group order, separated by blank lines:

```go
import (
    // 1. Standard library
    "context"
    "fmt"
    "time"

    // 2. Internal project imports
    "github.com/axiomod/axiomod/framework/config"
    "github.com/axiomod/axiomod/platform/observability"

    // 3. Third-party imports
    "github.com/gofiber/fiber/v2"
    "go.uber.org/fx"
    "go.uber.org/zap"
)
```

Use import aliases only to resolve conflicts:

```go
grpc_pkg "github.com/axiomod/axiomod/framework/grpc"
grpc_go "google.golang.org/grpc"
```

## Constructor Pattern

```go
// Simple (no error possible)
func NewExampleMemoryRepository() *ExampleMemoryRepository {
    return &ExampleMemoryRepository{examples: make(map[string]*entity.Example)}
}

// With dependencies
func NewExampleHandler(
    createUseCase *usecase.CreateExampleUseCase,
    getUseCase    *usecase.GetExampleUseCase,
    logger        *observability.Logger,
) *ExampleHandler { ... }

// Can fail
func NewPluginRegistry(cfg *config.Config, logger *observability.Logger,
    metrics *observability.Metrics, health *health.Health) (*PluginRegistry, error) { ... }
```

## Documentation

Every exported type, function, and method MUST have a Go doc comment:

```go
// Plugin defines the interface that all plugins must implement.
type Plugin interface { ... }

// NewPluginRegistry creates a new plugin registry.
func NewPluginRegistry(...) (*PluginRegistry, error) { ... }
```

## Logging

Always use `observability.Logger` (wrapping zap) with structured fields:

```go
logger.Info("HTTP request",
    zap.String("method", method),
    zap.Int("status", status),
    zap.Duration("latency", latency),
)
logger.Error("Failed to create example", zap.Error(err))
```

**Never** use `fmt.Println` or `log.Println`.
