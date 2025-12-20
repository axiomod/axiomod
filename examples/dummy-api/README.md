# Examples/Dummy-Api

A Go Macroservice application built with the Axiomod framework.

## Features

- Modular architecture with clean separation of concerns
- Dependency injection using Uber FX
- Observability with logging, metrics, and tracing
- Pluggable components for flexibility
- HTTP and gRPC API support
- Comprehensive testing

## Getting Started

### Prerequisites

- Go 1.21+
- Docker (optional, for containerization and dependencies like Postgres/Jaeger)

### Building

```bash
# Build the project
go build -o bin/examples/dummy-api ./cmd/examples/dummy-api
```

### Running

```bash
# Ensure dependencies (e.g., PostgreSQL, Jaeger) are running

# Run with default configuration
./bin/examples/dummy-api

# Run with custom configuration
./bin/examples/dummy-api --config=path/to/config.yaml
```

## Project Structure

The project follows a clean architecture approach with the following structure:

- `cmd/examples/dummy-api`: Application entry point
- `internal/domain`: Business entities and rules
- `internal/usecase`: Application-specific business rules
- `internal/infrastructure`: Implementation details (DB repositories, external APIs)
- `tests`: Unit and integration tests
- `docs`: Documentation
- `scripts`: Build and deployment scripts
- `migrations`: Database migration files

## License

This project is licensed under the MIT License - see the LICENSE file for details.
