# Axiomod Framework - Architecture Overview

## Introduction

The Axiomod Framework is a modular, plugin-based framework designed for building scalable Go applications. It follows clean architecture principles to ensure separation of concerns, testability, and maintainability.

## Core Architecture Principles

### Clean Architecture

The framework is built on the principles of Clean Architecture, which emphasizes:

1. **Independence from frameworks**: The business logic is independent of external frameworks
2. **Testability**: Business rules can be tested without UI, database, web server, or any external element
3. **Independence from UI**: The UI can change without changing the business rules
4. **Independence from database**: The business rules are not bound to a specific database
5. **Independence from external agencies**: Business rules don't know anything about the outside world

### Dependency Injection

The framework uses Uber's FX library for dependency injection, which provides:

- Constructor-based dependency injection
- Lifecycle management for graceful startup and shutdown
- Module system for organizing dependencies
- Dependency visualization for debugging

### Plugin System

The plugin system allows for extending the framework's functionality without modifying its core. Plugins can be:

- Enabled/disabled via configuration
- Loaded dynamically at runtime
- Configured via the central configuration system

## Directory Structure

```plaintext
/axiomod
├── cmd/                                       # Application entrypoints
│   └── axiomod-server/                        # Main server app
│
├── framework/                             # Core framework components
│   ├── config/                            # Configuration management
│   ├── auth/                              # Authentication & OIDC
│   ├── database/                          # Database abstraction
│   └── worker/                            # Background worker pools
│
├── plugins/                               # Dynamic plugins
│
├── platform/                               # Shared infrastructure
│   ├── server/                            # HTTP & gRPC server setup
│   └── observability/                     # Logging, Metrics, Tracing
│
├── internal/                                  # Private application code
│
├── examples/                                  # Example implementations
│
├── scripts/                                   # Automation scripts
│
├── docs/                                      # Documentation
│
└── api/                                       # API contracts
```

## Key Components

### Configuration System

The configuration system uses Viper to load configuration from:

- YAML/JSON files
- Environment variables
- Command-line flags

Configuration is hierarchical and can be overridden at different levels.

### Observability Platform

The observability platform provides:

- **Logging**: Structured logging using Uber's Zap
- **Metrics**: Metrics collection using Prometheus
- **Tracing**: Distributed tracing using OpenTelemetry

### HTTP/gRPC Servers

The framework supports both HTTP and gRPC protocols:

- HTTP server using Fiber v2
- gRPC server with interceptors for authentication, logging, etc.

### Database Access

Database access is abstracted through repositories:

- Support for MySQL and PostgreSQL
- Connection pooling and transaction management
- Connection pooling and transaction management

### Authentication & Authorization

The framework provides:

- JWT-based authentication
- Integration with Keycloak and OIDC Providers
- Role-based access control (RBAC) via JWT Claims

### Event System

The event system allows for asynchronous communication:

- Publish/subscribe pattern
- Support for Kafka and other message brokers
- Event-driven architecture

## Flow of Control

1. **Request Entry**: Requests enter through HTTP or gRPC servers
2. **Middleware Processing**: Requests pass through middleware for authentication, logging, etc.
3. **Handler/Controller**: Requests are routed to the appropriate handler
4. **Use Case Execution**: Business logic is executed in use cases
5. **Repository Access**: Data is retrieved or stored through repositories
6. **Response**: Results are returned to the client

## Extension Points

The framework can be extended through:

1. **Plugins**: Adding new functionality through the plugin system
2. **Middleware**: Adding request processing logic
3. **Repositories**: Implementing new data storage mechanisms
4. **Use Cases**: Implementing new business logic
5. **Event Handlers**: Processing events asynchronously

## Deployment Model

The framework supports:

- Containerization with Docker
- Kubernetes deployment
- CI/CD with GitLab CI
- Configuration through environment variables for different environments

## Security Considerations

- Secure defaults for all components
- Input validation to prevent injection attacks
- Authentication and authorization for all endpoints
- Audit logging for compliance
- Dependency scanning for vulnerabilities

## Performance Considerations

- Connection pooling for databases
- Caching for frequently accessed data
- Circuit breakers for resilience
- Timeouts and retries for external services
- Graceful shutdown for in-flight requests
