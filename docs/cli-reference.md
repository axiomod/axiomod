# Axiomod CLI Reference

The `axiomod` CLI is the primary tool for developing, managing, and deploying applications built with the Axiomod framework.

## Installation

```bash
make build-cli
# The binary will be available at ./bin/axiomod
```

## Global Flags

- `--config`: Path to the configuration file (default: `$HOME/.axiomod.yaml`).
- `--help`: Show help for command.

## Core Commands

### `init`

Initialize a new Axiomod project with the recommended Clean Architecture structure.

```bash
axiomod init [project-name]
```

**Creates:**

- Directory structure (`cmd`, `internal`, `pkg`, `docs`, etc.)
- `go.mod`
- Default configuration files
- `Makefile` and `Dockerfile`

### `config`

Manage configuration settings.

```bash
axiomod config view   # Display current configuration
axiomod config check  # Validate configuration file
```

### `version`

Display version information for the CLI and Framework.

## Code Generation (`generate`)

Scaffold new components to speed up development.

### `module`

Generate a new module structure.

```bash
axiomod generate module --name=order
```

### `service`

Generate a new service layer.

```bash
axiomod generate service --name=PaymentProcessor --module=billing
```

### `handler`

Generate HTTP or gRPC handlers.

```bash
axiomod generate handler --name=GetOrder --type=http --module=order
```

## Database Migrations (`migrate`)

Manage database schema changes safely.

### `create`

Create a new migration file pair (up/down).

```bash
axiomod migrate create add_users_table
```

### `up`

Apply all pending migrations.

```bash
axiomod migrate up
```

### `down`

Rollback the last applied migration.

```bash
axiomod migrate down
```

## Development Workflow

### `build`

Compile the application binary.

```bash
axiomod build
```

### `test`

Run unit and integration tests.

```bash
axiomod test
axiomod test --unit  # Run only unit tests
axiomod test --integration # Run only integration tests
```

### `lint`

Run configured linters (golangci-lint).

```bash
axiomod lint
```

### `fmt`

Format code using standard Go tools.

```bash
axiomod fmt
```

### `logs`

Tail application logs.

```bash
axiomod logs --follow
```

## DevOps & Deployment

### `dockerize`

Generate a Dockerfile and build the container image.

```bash
axiomod dockerize --tag=v1.0.0
```

### `deploy`

Deploy the application (requires provider config).

```bash
axiomod deploy --env=production
```

### `status`

Check the status of the running service.

```bash
axiomod status
```

### `healthcheck`

Ping the application's liveness and readiness probes.

```bash
axiomod healthcheck
```

## Plugins (`plugin`)

Manage extensions to the framework.

```bash
axiomod plugin list
axiomod plugin install [plugin-name]
axiomod plugin remove [plugin-name]
```

## Validator (`validator`)

Enforce architectural rules and code quality.

### `architecture`

Validate adherence to Clean Architecture dependency rules.

```bash
axiomod validator architecture
```

### `naming`

Check naming conventions.

```bash
axiomod validator naming
```

### `api-spec`

Validate OpenAPI/Swagger specifications.

```bash
axiomod validator api-spec
```

### `security`

Run security checks (gosec).

```bash
axiomod validator security
```
