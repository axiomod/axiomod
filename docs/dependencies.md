# Dependencies

This document lists the major dependencies used in Axiomod and the rationale for their selection.

## Core Framework

- **[go.uber.org/fx](https://github.com/uber-go/fx)**: Dependency injection framework. Chosen for its module system and lifecycle management, essential for building loosely coupled, testable applications.
- **[github.com/spf13/viper](https://github.com/spf13/viper)**: Configuration management. Industry standard for handling configuration from files, environment variables, and flags. Supports watching for changes.
- **[go.uber.org/zap](https://github.com/uber-go/zap)**: Structured logging. Selected for its high performance and structured output capabilities.

## Transports

- **[github.com/gofiber/fiber](https://github.com/gofiber/fiber)**: HTTP web framework. Chosen for its extreme performance (based on Fasthttp) and ease of use (Express-like API).
- **[google.golang.org/grpc](https://github.com/grpc/grpc-go)**: gRPC framework. Standard for high-performance inter-service communication.

## Persistence

- **[github.com/lib/pq](https://github.com/lib/pq)**: PostgreSQL driver. Standard Go driver for Postgres.
- **[database/sql](https://pkg.go.dev/database/sql)**: Standard library interface used for database access to ensure interchangeable drivers.

## Observability

- **[go.opentelemetry.io/otel](https://github.com/open-telemetry/opentelemetry-go)**: Tracing and metrics standard.
- **[github.com/prometheus/client_golang](https://github.com/prometheus/client_golang)**: Prometheus metrics client.

## Authentication & Authorization

- **[github.com/golang-jwt/jwt](https://github.com/golang-jwt/jwt)**: JWT implementation.
- **[github.com/casbin/casbin](https://github.com/casbin/casbin)**: Authorization library. Supports access control models like ACL, RBAC, ABAC.

## Messaging

- **[github.com/IBM/sarama](https://github.com/IBM/sarama)**: Kafka client. Mature and feature-rich library for Apache Kafka.

## Validation

- **[github.com/go-playground/validator](https://github.com/go-playground/validator)**: Struct validation. widely used for input validation.

## Development

- **[github.com/stretchr/testify](https://github.com/stretchr/testify)**: Testing toolkit. Provides assertions and mocks.
