# v1.3.0 - Developer Experience & Observability 🚀

This release brings major enhancements to observability, background processing, data layer capabilities, and developer tooling, creating a production-ready ecosystem.

## 🌟 Key Highlights

### 👁️ Observability (Epic P1.1)

- **Metrics Endpoint**: Native Prometheus metrics exposed at `/metrics`.
- **Distributed Tracing**: Full OpenTelemetry support with OTLP and Jaeger exporters.
- **Instrumented Middleware**: Automatic metrics and tracing for HTTP (Fiber) and gRPC requests.
- **K8s Health Checks**: Dedicated `/live` and `/ready` endpoints for accurate Kubernetes probing.

### 📨 Messaging & Background Jobs (Epic P1.4)

- **Kafka Fx Integration**: First-class support for Kafka Producers and Consumers with automatic lifecycle management (graceful start/stop).
- **Worker Pools**: Robust background worker integration with Fx, supporting job registration and graceful shutdown.

### 💾 Data Layer Enhancements (Epic P1.3)

- **Connection Pooling**: Configurable `MaxOpenConns`, `MaxIdleConns`, and `ConnMaxLifetime` for database stability.
- **Slow Query Logging**: Automatic logging of queries exceeding configurable thresholds.
- **Schema Migrations**: Built-in CLI commands for managing database schema via `golang-migrate`.
  - `beta migrate create`
  - `beta migrate up/down`
  - `beta migrate force/version`

### 🛠️ Developer Tooling (Epic P1.5)

- **CLI Enhancements**: Rebuilt CLI with comprehensive flag parsing fixes.
- **Code Generation**: New and improved generators for rapid development:
  - `generate service` (supports `--module` flag)
  - `generate handler` (supports `--module` flag)
  - `generate module` (creates full Clean Architecture structure)
- **Policy Management**: New `policy` command group for managing RBAC rules via CLI.

### 🔐 Auth & Security (Epic P1.2)

- **OIDC Optimization**: Discovery results are now cached to reduce latency and network calls.

## ⚠️ Breaking Changes

- **Plugin Interface**: The `Plugin` interface `Initialize` method now accepts `*observability.Metrics` and `*config.Config` as arguments. Custom plugins must be updated to match the new signature.
- **Database Wrapper**: `database.New` now requires `*observability.Metrics` and `*config.Config`.

## 📚 Updated Documentation

- [Observability Guide](./observability-guide.md): Configuration for metrics, OTLP, and health checks.
- [Events & Messaging Guide](./events-messaging-guide.md): Guide for using the new Kafka and Worker Fx modules.
- [Database Guide](./database-guide.md): Connection pooling and migration guide.
- [CLI Reference](./cli-reference.md): Complete reference for all new commands (`migrate`, `policy`, `generate`).
- [Deployment Guide](./deployment-guide.md): Updated Kubernetes manifests for liveness/readiness probes.
