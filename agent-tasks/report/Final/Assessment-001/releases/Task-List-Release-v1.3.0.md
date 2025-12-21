# Task List - Release v1.3.0

# P1: High Priority

## Epic P1.1: Observability

### Feature P1.1.1: Metrics Endpoint & Instrumentation

**Goal:** Expose Prometheus metrics and instrument HTTP/gRPC requests.

**Rationale:** The metrics registry exists but is never exposed. Metrics are essential for monitoring and alerting in production. Without metrics, operators are blind to system behavior.

**Scope:**

*In:*

- Create HTTP endpoint (e.g., `/metrics`) that serves the Prometheus registry.

- Implement Fiber middleware to record HTTP request metrics (count, duration, status code).

- Implement gRPC interceptor to record gRPC request metrics.

- Support configurable metrics port and path.

- Add custom metrics for application-specific events (e.g., job processed, invoice created).

- Implement histogram buckets for request latency.

*Out:*

- Metrics dashboards (future task).

- Alerting rules (future task).

**Dependencies:**

- `github.com/prometheus/client_golang` (already present).

**Risks:**

- Metrics recording could impact performance. *Mitigation:* Benchmark before/after; use efficient metric types.

- High cardinality metrics could exhaust memory. *Mitigation:* Use label filtering and aggregation.

**Acceptance Criteria:**

- [ ] `/metrics` endpoint is exposed on the configured port.

- [ ] HTTP request metrics are recorded (count, duration, status).

- [ ] gRPC request metrics are recorded.

- [ ] Metrics are queryable by Prometheus.

- [ ] Unit tests verify metric recording.

- [ ] Integration test with Prometheus scraping.

- [ ] Performance test shows <5% overhead.

**Test Plan:**

- Unit tests for middleware/interceptor.

- Integration test with Prometheus.

- Performance test to measure overhead.

**Documentation Needs:**

- Update `docs/observability-guide.md` with metrics configuration.

- Provide example Prometheus scrape config.

- Document available metrics and their meanings.

- Add troubleshooting guide for metrics issues.

**Release Notes Entry:**

```
## Observability
- **NEW:** Prometheus metrics endpoint is now exposed at `/metrics`. HTTP and gRPC requests are automatically instrumented.
- **NEW:** Custom application metrics can be easily registered and recorded.
```


---

### Feature P1.1.2: Distributed Tracing (OpenTelemetry)

**Goal:** Implement OpenTelemetry exporters for Jaeger and OTLP.

**Rationale:** Tracing is hardcoded to use a no-op exporter, making distributed tracing impossible. Distributed tracing is critical for debugging issues in microservices architectures.

**Scope:**

*In:*

- Implement Jaeger exporter configuration.

- Implement OTLP exporter configuration.

- Create Fiber middleware to start spans for HTTP requests.

- Create gRPC interceptor to start spans for gRPC requests.

- Propagate trace context to downstream services.

- Support configurable sampling ratio.

- Implement trace context propagation in Kafka messages.

*Out:*

- Custom span attributes (future task).

- Trace sampling strategies (future task).

**Dependencies:**

- `go.opentelemetry.io/otel` (already present).

- `go.opentelemetry.io/otel/exporters/jaeger/otlphttp` or `otlptracegrpc`.

**Risks:**

- Exporter configuration complexity. *Mitigation:* Provide sensible defaults and examples.

- Performance impact of tracing. *Mitigation:* Implement sampling and benchmark.

**Acceptance Criteria:**

- [ ] Jaeger exporter is functional.

- [ ] OTLP exporter is functional.

- [ ] HTTP requests generate spans.

- [ ] gRPC requests generate spans.

- [ ] Trace context is propagated.

- [ ] Unit tests verify span creation.

- [ ] Integration test with Jaeger.

- [ ] Performance test shows <10% overhead with sampling.

**Test Plan:**

- Unit tests for exporter initialization.

- Integration test with Jaeger or Jaeger-in-a-box.

- Verify trace context propagation.

**Documentation Needs:**

- Update `docs/observability-guide.md` with tracing configuration.

- Provide example Jaeger/OTLP configurations.

- Document trace context propagation.

- Add troubleshooting guide for tracing issues.

**Release Notes Entry:**

```
## Observability
- **NEW:** Distributed tracing via Jaeger or OTLP. Configure `observability.tracingExporterType` and `observability.tracingExporterURL` in your config.
- **NEW:** Trace context is automatically propagated across services.
```


---

### Feature P1.1.3: Health Check Endpoints

**Goal:** Register `/live` and `/ready` endpoints for Kubernetes probes.

**Rationale:** Health checks are implemented but not exposed as HTTP endpoints. Kubernetes deployments require these for pod readiness and liveness.

**Scope:**

*In:*

- Register `/live` endpoint (liveness probe ).

- Register `/ready` endpoint (readiness probe).

- Implement readiness checks for database, Kafka, and other dependencies.

- Support custom readiness hooks.

- Implement health check details endpoint.

*Out:*

- Detailed health status reporting (future task).

**Dependencies:**

- None (internal task).

**Risks:**

- Readiness checks could be slow. *Mitigation:* Cache results with a short TTL.

- Cascading failures if dependencies are down. *Mitigation:* Implement graceful degradation.

**Acceptance Criteria:**

- [ ] `/live` endpoint returns 200 if the service is running.

- [ ] `/ready` endpoint returns 200 if all dependencies are healthy.

- [ ] Database connectivity is checked in readiness.

- [ ] Kafka connectivity is checked in readiness (if enabled).

- [ ] Custom readiness hooks can be registered.

- [ ] Unit tests verify endpoint behavior.

- [ ] Integration test with dependencies.

**Test Plan:**

- Unit tests for health checks.

- Integration test with dependencies.

**Documentation Needs:**

- Update `docs/deployment-guide.md` with Kubernetes probe configuration.

- Provide example Kubernetes manifests.

- Document custom health check registration.

**Release Notes Entry:**

```
## Observability
- **NEW:** Health check endpoints `/live` and `/ready` are now available for Kubernetes probes.
- **NEW:** Custom health checks can be registered for application-specific dependencies.
```



---

## Epic P1.2: Authentication & Authorization

### Feature P1.2.1: OIDC Discovery Caching

**Goal:** Cache OIDC discovery results to avoid repeated network calls.

**Rationale:** OIDC discovery is performed on every token verification, which is inefficient and could be a security risk.

**Scope:**

*In:*

- Implement caching of OIDC discovery results with configurable TTL.

- Implement automatic refresh of cached results.

- Add error handling for discovery failures.

- Support manual refresh via API.

*Out:*

- Multi-issuer caching (future task).

**Dependencies:**

- None (internal task).

**Risks:**

- Stale discovery results could cause issues. *Mitigation:* Implement refresh logic and error handling.

**Acceptance Criteria:**

- [ ] Discovery results are cached with configurable TTL (default 1 hour).

- [ ] Cache is automatically refreshed when expired.

- [ ] Manual refresh is available via API.

- [ ] Error handling for discovery failures.

- [ ] Unit tests verify caching behavior.

**Test Plan:**

- Unit tests for caching logic.

- Integration test with mock OIDC provider.

**Documentation Needs:**

- Update `docs/auth-security-guide.md` with caching configuration.

**Release Notes Entry:**

```
## Performance
- **IMPROVED:** OIDC discovery results are now cached to reduce network calls.
```

---

## Epic P1.3: Data Layer

### Feature P1.3.1: Database Plugin Enhancement

**Goal:** Enhance database plugins with connection pool configuration and instrumentation.

**Rationale:** Database plugins are minimal and lack configuration options for production use.

**Scope:**

*In:*

- Add connection pool settings (max open/idle connections, connection lifetime).

- Add slow query logging.

- Add query duration metrics.

- Support for read/write splitting (configuration only, not implementation).



*Out:*

- Full ORM implementation (future task).

**Dependencies:**

- None (internal task).

**Risks:**

- Connection pool misconfiguration could cause issues. *Mitigation:* Provide sensible defaults and documentation.

**Acceptance Criteria:**

- [ ] Connection pool settings are configurable.

- [ ] Slow queries are logged.

- [ ] Query duration metrics are recorded.

- [ ] Unit tests verify configuration.

- [ ] Integration test with real database.

**Test Plan:**

- Unit tests for configuration.

- Integration test with MySQL/PostgreSQL.

**Documentation Needs:**

- Update `docs/database-guide.md` with connection pool configuration.

- Provide tuning recommendations.

**Release Notes Entry:**

```
## Data Layer
- **IMPROVED:** Database plugins now support connection pool configuration and slow query logging.
```


---

### Feature P1.3.2: Database Migration Support

**Goal:** Implement CLI commands for database migrations.

**Rationale:** The CLI has migration commands defined but they are non-functional.

**Scope:**

*In:*

- Integrate `golang-migrate` or similar tool.

- Implement `axiomod migrate create` command.

- Implement `axiomod migrate up` command.

- Implement `axiomod migrate down` command.

- Provide migration templates.

*Out:*

- Migration versioning strategies (future task).

**Dependencies:**

- `github.com/golang-migrate/migrate/v4`.

**Risks:**

- Migration failures could corrupt data. *Mitigation:* Require explicit confirmation; provide rollback guidance.

**Acceptance Criteria:**

- [ ] `axiomod migrate create` generates migration files.

- [ ] `axiomod migrate up` applies pending migrations.

- [ ] `axiomod migrate down` rolls back the last migration.

- [ ] Migration status can be queried.

- [ ] Unit tests verify command behavior.

**Test Plan:**

- Unit tests for CLI commands.

- Integration test with real database.

**Documentation Needs:**

- Add `docs/migration-guide.md` with migration patterns.

- Provide example migrations.

**Release Notes Entry:**

```
## CLI
- **NEW:** Database migration commands are now available: `axiomod migrate create`, `axiomod migrate up`, `axiomod migrate down`.
```

---

## Epic P1.4: Messaging & Background Jobs

### Feature P1.4.1: Kafka DI Integration

**Goal:** Wire Kafka producers and consumers into the dependency injection container.

**Rationale:** Kafka is implemented but not integrated with DI, making it cumbersome to use.

**Scope:**

*In:*

- Create Fx module for Kafka producer.

- Create Fx module for Kafka consumer.

- Support consumer group configuration.

- Support handler registration via DI.

- Implement graceful shutdown for consumers.

*Out:*

- Dead-letter queue support (future task).

- Exactly-once semantics (future task).

**Dependencies:**

- `github.com/Shopify/sarama` (already present).

**Risks:**

- Consumer group coordination could be complex. *Mitigation:* Use Sarama's built-in group coordination.

**Acceptance Criteria:**

- [ ] Kafka producer is provided via DI.

- [ ] Kafka consumer is provided via DI.

- [ ] Consumer handlers are registered via DI.

- [ ] Graceful shutdown works for consumers.

- [ ] Unit tests verify DI wiring.

- [ ] Integration test with real Kafka broker.

**Test Plan:**

- Unit tests for DI modules.

- Integration test with Docker Compose Kafka.

**Documentation Needs:**

- Update `docs/events-messaging-guide.md` with DI integration examples.

- Provide example producer/consumer code.

**Release Notes Entry:**

```
## Messaging
- **IMPROVED:** Kafka producers and consumers are now integrated with dependency injection for easier configuration.
```


---

### Feature P1.4.2: Worker Pool Integration

**Goal:** Wire background worker pools into the dependency injection container.

**Rationale:** Workers are implemented but not integrated with DI.

**Scope:**

*In:*

- Create Fx module for worker pool.

- Support job registration via DI.

- Support graceful shutdown.

- Implement job scheduling (cron-like).

*Out:*

- Persistent job queues (future task).

**Dependencies:**

- None (internal task).

**Risks:**

- Job scheduling complexity. *Mitigation:* Use a simple cron-like syntax.

**Acceptance Criteria:**

- [ ] Worker pool is provided via DI.

- [ ] Jobs are registered via DI.

- [ ] Graceful shutdown works for workers.

- [ ] Job scheduling is supported.

- [ ] Unit tests verify DI wiring.

**Test Plan:**

- Unit tests for DI modules.

- Integration test with sample jobs.

**Documentation Needs:**

- Update `docs/events-messaging-guide.md` with worker examples.

- Provide example job code.

**Release Notes Entry:**

```
## Background Jobs
- **IMPROVED:** Worker pools are now integrated with dependency injection for easier configuration.
```


---

## Epic P1.5: Developer Tooling

### Feature P1.5.1: CLI Rebuild & Testing

**Goal:** Rebuild the CLI binary to include all commands and fix flag recognition.

**Rationale:** The compiled CLI binary is outdated and does not recognize flags for `generate` and other commands.

**Scope:**

*In:*

- Rebuild the CLI binary with all commands.

- Fix flag parsing for `generate`, `migrate`, and `validator` commands.

- Add automated tests for CLI commands.

- Update CLI help text.

*Out:*

- New CLI commands (future task).

**Dependencies:**

- None (internal task).

**Risks:**

- CLI regressions. *Mitigation:* Add automated tests for each command.

**Acceptance Criteria:**

- [ ] CLI binary includes all commands.

- [ ] Flags are recognized correctly.

- [ ] `axiomod generate module --name=foo` works.

- [ ] `axiomod migrate create` works.

- [ ] `axiomod validator run` works.

- [ ] Automated tests pass.

**Test Plan:**

- Unit tests for CLI commands.

- Integration test with real project.

**Documentation Needs:**

- Update `docs/cli-reference.md` with correct command syntax.

**Release Notes Entry:**

```
## CLI
- **FIXED:** CLI commands now recognize all flags correctly. Rebuild your binary with `make build-cli`.
```


---

### Feature P1.5.2: Code Generation Improvement

**Goal:** Improve code generation to create complete Clean Architecture slices.

**Rationale:** The `generate` commands are incomplete and don't create all necessary files.

**Scope:**

*In:*

- Update `generate module` to create entity, repository, usecase, service, delivery/http, delivery/grpc, infrastructure directories.

- Implement `generate service` command.

- Implement `generate handler` command.

- Add templates for common patterns.

*Out:*

- OpenAPI-first code generation (future task ).

**Dependencies:**

- None (internal task).

**Risks:**

- Generated code quality. *Mitigation:* Provide well-tested templates.

**Acceptance Criteria:**

- [ ] `generate module` creates all necessary directories.

- [ ] `generate service` creates service skeleton.

- [ ] `generate handler` creates handler skeleton.

- [ ] Generated code compiles and follows conventions.

- [ ] Unit tests verify code generation.

**Test Plan:**

- Unit tests for code generation.

- Integration test with generated code.

**Documentation Needs:**

- Update `docs/cli-reference.md` with code generation examples.

**Release Notes Entry:**

```
## CLI
- **IMPROVED:** Code generation now creates complete Clean Architecture slices with all necessary files and directories.
```

