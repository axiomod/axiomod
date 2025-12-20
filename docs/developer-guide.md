# Developer Guide

This guide provides information for developers who are building applications using the Axiomod framework.

## 1. Environment Setup

### Prerequisites

- **Go**: Version 1.24.2 or higher.
- **Docker**: For running external dependencies like MySQL, PostgreSQL, or Kafka.
- **golangci-lint**: For code linting.
- **Cobra CLI**: The framework uses Cobra for its CLI tool.

### Installation

1. Clone the repository:

   ```bash
   git clone https://github.com/axiomod/axiomod.git
   cd axiomod
   ```

2. Install dependencies:

   ```bash
   make deps
   ```

## 2. CLI Tool (`axiomod`)

The `axiomod` CLI tool is the primary way to interact with the framework.

### Building the CLI

```bash
make build-cli
```

The binary will be available at `./bin/axiomod`.

### Common Commands

- **New Project**: Initialize a new project structure.

  ```bash
  ./bin/axiomod init my-new-service
  ```

  This command generates the basic directory structure, `go.mod` (with framework `replace` directive if in dev mode), default configuration, and a `main.go` entry point.

- **Migrations**: Manage database migrations.

  ```bash
  ./bin/axiomod migrate create initial_schema
  ./bin/axiomod migrate up
  ```

- **Validation**: Run code validators to ensure architecture and naming standards.

  ```bash
  ./bin/axiomod validator run
  ```

## 3. Development Workflow

The framework provides a `Makefile` in the `scripts/` directory (aliased to the root via scripts if applicable, or run directly).

### Standard Commands

- **Build**: Compiles the application.

  ```bash
  make build
  ```

- **Test**: Runs all unit tests.

  ```bash
  make test
  ```

- **Lint**: Runs `golangci-lint` to ensure code quality.

  ```bash
  make lint
  ```

- **Format**: Automatically formats Go code.

  ```bash
  make fmt
  ```

- **Code Generation**: Runs `go generate` for Ent or other generators.

  ```bash
  make generate
  ```

## 4. Project Structure

The framework follows Clean Architecture principles:

- `cmd/`: Application entry points (Main server and CLI).
- `framework/`: Core framework components (config, auth, database, etc.).
- `plugins/`: Extension points for the framework.
- `platform/`: Infrastructure-level code (server setup, observability).
- `internal/`: Private application-specific code.
- `examples/`: Sample implementations to guide your development.

## 5. Adding Functionality

### Implementing an HTTP Handler

1. **Create the Handler**: Use the `observability.Logger` wrapper for logging.

    ```go
    package http

    import (
        "github.com/axiomod/axiomod/platform/observability"
        "github.com/gofiber/fiber/v2"
    )

    type MyHandler struct {
        logger *observability.Logger
    }

    func NewMyHandler(logger *observability.Logger) *MyHandler {
        return &MyHandler{logger: logger}
    }

    func (h *MyHandler) RegisterRoutes(app *fiber.App) {
        app.Get("/my-endpoint", h.HandleRequest)
    }
    ```

2. **Define an Fx Module**: Create a module to provide the handler and register routes.

    ```go
    package internal

    import (
        "go.uber.org/fx"
        "my-service/internal/delivery/http"
        "github.com/axiomod/axiomod/platform/server"
    )

    var Module = fx.Module(
        "my_module",
        fx.Provide(http.NewMyHandler),
        fx.Invoke(registerRoutes),
    )

    func registerRoutes(handler *http.MyHandler, s *server.HTTPServer) {
        handler.RegisterRoutes(s.App)
    }
    ```

3. **Wire into `main.go`**: Add your module to the `fx.New` list and ensure servers are registered.

    ```go
    app := fx.New(
        // ... platform modules
        internal.Module,

        // CRITICAL: Register HTTP/gRPC servers
        fx.Invoke(
            server.RegisterHTTPServer,
            server.RegisterGRPCServer,
        ),
    )
    ```

### Adding a Plugin

Refer to the [Plugin Development Guide](./plugin-development-guide.md) for detailed instructions on creating and registering plugins.

## 6. Local Development

When developing the framework itself or testing against a local clone, use the `replace` directive in your project's `go.mod`:

```go
replace github.com/axiomod/axiomod => ../path/to/axiomod
```

This directs the Go toolchain to use your local source instead of fetching from the remote repository.

## 7. Build and Verification

### Building Locally

To keep your project clean, build binaries into a `bin/` directory:

```bash
mkdir -p bin
go build -o bin/my-service ./cmd/my-service/main.go
```

### Running the Service

Start the service with a configuration file:

```bash
./bin/my-service --config config/service_default.yaml
```

### Verifying Endpoints

Use `curl` to verify your endpoints sequentially:

```bash
# Framework default health checks
curl -s http://localhost:8080/live
curl -s http://localhost:8080/ready

# Your custom endpoints
curl -s http://localhost:8080/my-endpoint
```
