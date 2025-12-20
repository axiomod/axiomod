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
   cd axiomod-framework
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

- **New Project**: Initialize a new project from a template.

  ```bash
  ./bin/axiomod init my-new-service
  ```

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
- `internal/framework/`: Core framework components (config, auth, database, etc.).
- `internal/plugins/`: Extension points for the framework.
- `internal/platform/`: Infrastructure-level code (server setup, observability).
- `internal/examples/`: Sample implementations to guide your development.

## 5. Adding Functionality

### Adding a Use Case

1. Define the input/output ports (interfaces) in the domain layer.
2. Implement the use case logic in the use case layer.
3. Inject the use case into the delivery layer (HTTP/gRPC) using Fx.

### Adding a Plugin

Refer to the [Plugin Development Guide](./plugin-development-guide.md) for detailed instructions on creating and registering plugins.
