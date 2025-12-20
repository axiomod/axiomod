# Enterprise Axiomod: Enhancement Roadmap

While the current framework is production-ready, truly large-scale enterprise environments have specialized needs regarding multi-tenancy, compliance, and operational scale. This roadmap outlines detailed enhancements to address those needs.

## 1. Multi-Tenancy Support (High Priority)

**Goal**: Native support for serving multiple tenants (clients/customers) securely from a single deployment.

- **Tenant Context Middleware**:
  - Automatically extract `Tenant-ID` from headers or subdomains.
  - Inject tenant context into `context.Context` for pervasive access.
- **Data Isolation Strategies**:
  - **Row-Level**: Automatically applying `WHERE tenant_id = ?` clauses via repository helpers.
  - **Schema-Based**: Utilizing Postgres schemas (`SET search_path`) per tenant request.
  - **Database-Based**: Dynamic connection switching based on tenant context.
- **Tenant Config**: Ability to override specific configurations (e.g., feature flags, rate limits) on a per-tenant basis.

## 2. Advanced Security & Compliance

**Goal**: Meet strict banking/healthcare/fintech security standards.

- **Vault Integration**:
  - Native provider for HashiCorp Vault to fetch secrets at runtime (replacing static env vars).
  - Automatic lease renewal and rotation handling.
- **Audit Logging 2.0**:
  - Structured, tamper-evident audit logs separate from application logs.
  - Contextual capture: `ActorID`, `Action`, `Resource`, `TenantID`, `OldValue`, `NewValue`.
  - Async shipping to immutable storage (e.g., S3 WORM, Splunk).
- **Rate Limiting & Throttling**:
  - Distributed rate limiting using Redis (Sliding Window or Token Bucket).
  - Per-user, per-IP, and per-tenant limits.
- **mTLS Support**:
  - Native flags to enforce mutual TLS for service-to-service gRPC communication.

## 3. Resilience & Stability

**Goal**: ensuring the system survives failure of dependencies.

- **Advanced Circuit Breakers**:
  - Integration with libraries like `gobreaker` or resilience4j patterns.
  - Configurable failure thresholds per external client (not one global setting).
- **Bulkheading**:
  - Limiting the max concurrent requests for specific heavy endpoints to prevent cascading failures.
- **Adaptive Concurrency**:
  - Automatically shedding load when CPU/Memory reaches critical thresholds (Active Queue Management).
- **Graceful Degradation**:
  - Fallback mechanisms for critical read paths (e.g., serve stale cache if DB is down).

## 4. Developer Experience & Tooling

**Goal**: Accelerate adoption and reduce boilerplate.

- **Enhanced CLI Code Generation**:
  - `axiomod generate scaffold`: Generate a full slice (Entity, DTOs, Handler, Service, Repository, Tests) from a schema definition.
  - Support for OpenAPI-first generation (Spec -> Code).
- **Architecture Validation CI**:
  - A linter plugin that strictly adheres to dependency rules (e.g., Domain cannot import Driver).
- **Contract Testing**:
  - Integration with Pact for consumer-driven contract testing between microservices.

## 5. Observability Enhancements

**Goal**: Faster MTTR (Mean Time To Recovery).

- **Distributed Context Propagation**:
  - Full support for OpenTelemetry Baggage to propagate business context (User ID, Transaction ID) across service boundaries alongside trace IDs.
- **SLO/SLI Metrics**:
  - Built-in middleware to measure specific Service Level Indicators (latency success rate) and alert on error budgets.
- **Profiling Endpoints**:
  - Safe, protected `pprof` endpoints enabled via configuration for on-demand production profiling.

## 6. Event Driven Architecture (Advanced)

**Goal**: Robust asynchronous flows.

- **Outbox Pattern**:
  - Native implementation of the Transactional Outbox pattern to allow atomic "Save to DB + Publish Event" operations.
- **Dead Letter Queues (DLQ)**:
  - Automatic routing of failed messages to DLQ topics with CLI tools to replay them.
