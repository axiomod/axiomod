# Axiomod Framework: Core Architecture

This document describes the core architecture, dependency management, and component lifecycle of the Axiomod framework.

## 1. Dependency Injection (Uber Fx)

The framework uses [Uber Fx](https://github.com/uber-go/fx) for dependency injection. This allows for a modular design where each component provides its own `fx.Option` (Module).

### Application Entrypoint

The main entrypoint is in `cmd/axiomod-server/main.go`. it bootstraps the application by:

1. Loading the configuration.
2. Registering all modules defined in `fx_options.go`.
3. Starting the Fx application.

### Defining a Module

Every framework component should provide a `Module` variable:

```go
package mycomponent

import "go.uber.org/fx"

var Module = fx.Options(
    fx.Provide(NewService),
    fx.Invoke(RegisterLifecycle),
)
```

## 2. Component Lifecycle

The framework manages the lifecycle of components using Fx hooks.

- **OnStart:** Used to start servers (HTTP, gRPC), connect to databases, or start background workers.
- **OnStop:** Used for graceful shutdown, closing database connections, and Stopping workers.

### Built-in Lifecycles

- **HTTP Server:** Managed in `platform/server`. Starts on a configured port and shuts down gracefully.
- **Plugin Registry:** Managed in `plugins`. Automatically starts and stops all enabled plugins.
- **Background Worker:** Managed in `framework/worker`. Ensures all jobs are stopped on application shutdown.

## 3. Error Handling

A standard error handling mechanism is provided in `framework/errors`.

### Guidelines

- Use sentinel errors for common cases (e.g., `errors.ErrNotFound`).
- Wrap errors with context using `errors.Wrap(err, "context message")`.
- Use helpers like `errors.NewInternal(err, "message")` to automatically assign error codes.
- Error codes are mapped to HTTP and gRPC status codes automatically in delivery layers.

## 4. How-To: Adding a New Plugin

1. Create a new struct implementing the `Plugin` interface in `plugins`.
2. Implement `Name()`, `Initialize()`, `Start()`, and `Stop()`.
3. Register the plugin in `PluginRegistry` (currently in `builtin_plugins.go`).
4. Enable the plugin in `service_default.yaml`.
