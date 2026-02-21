# Observability

## Logger

`observability.Logger` wraps `*zap.Logger`:

```go
logger.Info("message", zap.String("key", "value"))
logger.Error("failed", zap.Error(err))
logger.Warn("warning", zap.String("detail", detail))
logger.Debug("debug info", zap.Int("count", n))
```

- Configured from `ObservabilityConfig.LogLevel` and `LogFormat`
- Adds service name and environment as default fields
- **Never** use `fmt.Println`, `log.Println`, or raw `zap.L()`

## Tracer

`observability.Tracer` wraps OpenTelemetry:

```go
ctx, span := tracer.Tracer.Start(ctx, "operation-name")
defer span.End()

span.SetAttributes(attribute.String("key", "value"))
span.RecordError(err)
```

Supported exporters: `jaeger`, `otlp`, `stdout`. Configurable sampling ratio.

## Metrics

`observability.Metrics` wraps Prometheus:

```go
metrics.HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()
metrics.HTTPRequestDuration.WithLabelValues(method, path, status).Observe(duration)
metrics.GRPCRequestsTotal.WithLabelValues(service, method, status).Inc()
metrics.DBQueryDuration.WithLabelValues(queryType, status).Observe(duration)
```

Pre-defined metric vectors:
- `HTTPRequestsTotal` (CounterVec: method, path, status)
- `HTTPRequestDuration` (HistogramVec: method, path, status)
- `GRPCRequestsTotal` (CounterVec: service, method, status)
- `GRPCRequestDuration` (HistogramVec: service, method, status)
- `DBQueryDuration` (HistogramVec: query_type, status)

## Health Checks

Register checks for any component:

```go
health.RegisterCheck("database", func() error {
    return db.Ping()
})
```

Exposed at `/live`, `/ready`, `/health`.

## Rules

1. Use structured logging with `zap` fields, not string formatting
2. Create spans for significant operations (DB queries, external calls, use case execution)
3. Always `defer span.End()` after starting a span
4. Record errors on spans via `span.RecordError(err)`
5. Use pre-defined Prometheus metrics, add new vectors to `observability.Metrics` struct
6. Register health checks for all external dependencies
