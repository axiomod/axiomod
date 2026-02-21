# Middleware Patterns

## Struct-with-Handle Pattern

All middleware follows this structure:

```go
type LoggingMiddleware struct {
    logger *observability.Logger
}

func NewLoggingMiddleware(logger *observability.Logger) *LoggingMiddleware {
    return &LoggingMiddleware{logger: logger}
}

func (m *LoggingMiddleware) Handle() fiber.Handler {
    return func(c *fiber.Ctx) error {
        start := time.Now()
        err := c.Next()
        m.logger.Info("HTTP request",
            zap.String("method", c.Method()),
            zap.String("path", c.Path()),
            zap.Int("status", c.Response().StatusCode()),
            zap.Duration("latency", time.Since(start)),
        )
        return err
    }
}
```

## Available Middleware

| Middleware | Purpose |
|---|---|
| `LoggingMiddleware` | Logs method, path, status, latency, IP, user-agent |
| `AuthMiddleware` | JWT token extraction/validation, stores claims in `c.Locals` |
| `RoleMiddleware` | Checks `c.Locals("roles")` for required role |
| `TimeoutMiddleware` | Wraps request in `context.WithTimeout` |
| `RecoveryMiddleware` | Recovers from panics, returns 500 |
| `MetricsMiddleware` | Prometheus counters and histograms |
| `TracingMiddleware` | OTel spans with HTTP attributes |
| `RBACMiddleware` | Casbin-based policy enforcement |

## Application Order (server-level)

Applied in `platform/server/server.go`:

1. `recover.New()` (Fiber built-in)
2. `cors.New()` (Fiber built-in)
3. `compress.New()` (Fiber built-in)
4. `logger.New()` (Fiber built-in)
5. `metricsMid.Handle()` (custom)
6. `tracingMid.Handle()` (custom)

## Domain-Level Middleware

Applied at the route group level in `module.go`:

```go
api := app.Group("/api/v1")
api.Use(loggingMiddleware.Handle())
api.Use(authMiddleware.Handle())
handler.RegisterRoutes(api)
```

## RBAC Middleware

Takes resource and action as parameters:

```go
api.Use(rbacMiddleware.Handle("data1", "read"))
```

## Rules

1. All middleware uses the struct-with-`Handle()` pattern
2. `Handle()` returns `fiber.Handler`
3. Always call `c.Next()` to pass to the next handler
4. Server-level middleware in `platform/server/`, domain middleware at group level
5. Inject middleware dependencies via constructor (logger, metrics, etc.)
