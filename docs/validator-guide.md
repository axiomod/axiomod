# Validator Guide

## Introduction

The Axiomod framework includes a comprehensive set of validator tools to ensure code quality, adherence to architectural rules, naming conventions, and other standards. These validators help maintain consistency and quality across your codebase.

## Available Validators

The following validators are available through the `axiomod validator` command:

| Validator | Description |
|-----------|-------------|
| `architecture` | Validates that code follows the defined architectural dependencies |
| `naming` | Checks naming conventions for Go code, API endpoints, and database schemas |
| `domain` | Ensures domain boundaries are respected according to defined rules |
| `static-analysis` | Runs all static analysis tools (vet, gosec, staticcheck) |
| `static-check` | Runs staticcheck static analyzer |
| `security` | Runs gosec security scanner |
| `check-api-spec` | Checks API spec against standards using spectral |
| `check-docs` | Checks if code changes have documentation updates |
| `standards-check` | Runs all validators |
| `all` | Alias for standards-check |

## Architecture Validator

The architecture validator ensures that your code follows the defined architectural dependencies. It checks that modules only import from allowed dependencies as defined in the architecture rules.

### Usage

```bash
axiomod validator architecture [dir]
```

### Options

```
--config string   Path to architecture rules configuration file
```

### Architecture Rules

The architecture validator uses a JSON configuration file to define allowed dependencies between modules. By default, it looks for `architecture-rules.json` in the current directory, the `./config` directory, or the `./internal/framework/config` directory. You can also specify a custom path using the `--config` flag.

Example `architecture-rules.json`:

```json
{
  "allowedDependencies": {
    "entity": [],
    "repository": ["entity"],
    "usecase": ["entity", "repository", "service"],
    "service": ["entity", "repository"],
    "delivery/http": ["usecase", "entity", "middleware"],
    "delivery/grpc": ["usecase", "entity"],
    "infrastructure/persistence": ["entity", "repository"],
    "infrastructure/cache": ["entity"],
    "infrastructure/messaging": ["entity"],
    "platform/*": ["config"],
    "plugins/*": ["platform/*"]
  },
  "exceptions": [
    "vendor",
    "mocks",
    "test"
  ],
  "domainRules": {
    "allowCrossDomainDependencies": false,
    "allowedCrossDomainImports": [
      {
        "source": "internal/examples/example",
        "target": "internal/platform",
        "explanation": "Example domain can import platform components"
      }
    ]
  }
}
```

## Naming Validator

The naming validator checks that your code follows the project's naming conventions for Go code, API endpoints, database schemas, and more.

### Usage

```bash
axiomod validator naming [dir]
```

### Options

```
--sql string   Directory containing SQL migrations (default "./internal/platform/ent/migrate/migrations")
--api string   Directory containing API handlers (default "./internal/examples/example/delivery/http")
--json         Output results in JSON format
```

### Naming Conventions

The naming validator checks the following conventions:

- **Go Code**:
  - Package names: lowercase, single words
  - Exported functions/variables: PascalCase
  - Unexported functions/variables: camelCase
  - Types: PascalCase
  - Interface names: Often with -er suffix
  - File names: snake_case.go

- **API Endpoints**:
  - Lowercase with hyphens as separators
  - No trailing slashes (except for root endpoint)
  - Proper versioning format (/v{number}/)
  - Resource names should be plural for collections

- **Database**:
  - Table names: snake_case and plural
  - Column names: snake_case
  - Foreign key columns: end with _id
  - Boolean columns: prefix with is_, has_, can_, etc.
  - Timestamp columns: suffix with _at, _date, or _time

## Domain Validator

The domain validator ensures that domain boundaries are respected according to defined rules. It checks that imports between domains follow the allowed dependency rules.

### Usage

```bash
axiomod validator domain [dir]
```

### Options

```
--config string   Path to architecture rules configuration file
```

The domain validator uses the same configuration file as the architecture validator.

## Static Analysis Validators

### Static Analysis

Runs all static analysis tools (go vet, staticcheck, gosec).

```bash
axiomod validator static-analysis [dir]
```

### Static Check

Runs staticcheck static analyzer.

```bash
axiomod validator static-check [dir]
```

### Security

Runs gosec security scanner.

```bash
axiomod validator security [dir]
```

## API Spec Validator

Checks API specification files (OpenAPI/Swagger) against standards using spectral.

### Usage

```bash
axiomod validator check-api-spec [dir]
```

## Documentation Validator

Checks if code changes in the last commit have corresponding documentation updates.

### Usage

```bash
axiomod validator check-docs [dir]
```

## Running All Validators

You can run all validators at once using the `standards-check` or `all` command.

### Usage

```bash
axiomod validator standards-check [dir]
```

or

```bash
axiomod validator all [dir]
```

### Options

```
--config string   Path to architecture rules configuration file
--sql string      Directory containing SQL migrations
--api string      Directory containing API handlers
```

## Integration with CI/CD

You can integrate these validators into your CI/CD pipeline to ensure code quality and standards compliance.

Example GitLab CI configuration:

```yaml
validate:
  stage: test
  script:
    - go build -o bin/axiomod ./cmd/axiomod
    - ./bin/axiomod validator all
  only:
    - merge_requests
```

## Conclusion

The validator tools in Axiomod help maintain code quality and consistency across your project. By running these validators regularly, you can ensure that your code follows the defined architectural rules, naming conventions, and other standards.

## Writing Custom Validation Rules

You can extend the architecture validator by writing custom validation rules. Here's how to create and add new validation rules:

### 1. Understanding the Rule Structure

Architecture rules are defined in the `architecture-rules.json` file with the following structure:

```json
{
  "allowedDependencies": {
    "moduleA": ["moduleB", "moduleC"],
    "moduleD": ["moduleE"]
  },
  "exceptions": ["vendor", "mocks"],
  "domainRules": {
    "allowCrossDomainDependencies": false,
    "allowedCrossDomainImports": [
      {
        "source": "domainA",
        "target": "domainB",
        "explanation": "Reason for exception"
      }
    ]
  }
}
```

### 2. Adding New Module Dependencies

To add a new module and its allowed dependencies:

1. Identify the module name (e.g., "internal/newmodule")
2. Determine which other modules it should be allowed to import
3. Add an entry to the `allowedDependencies` section:

```json
"internal/newmodule": ["entity", "repository"]
```

### 3. Adding Cross-Domain Exceptions

If you need to allow specific cross-domain imports:

1. Set `allowCrossDomainDependencies` to `false` (to enforce domain boundaries)
2. Add specific exceptions to `allowedCrossDomainImports`:

```json
{
  "source": "internal/examples/newservice",
  "target": "internal/framework/logger",
  "explanation": "New service needs framework logging"
}
```

### 4. Testing Your Rules

After adding new rules, test them with:

```bash
axiomod validator architecture --config=path/to/your/architecture-rules.json
```

This will validate your codebase against the updated rules.



## Detailed CLI Usage Examples

Here are more detailed examples of how to use the `axiomod validator` subcommands:

### Architecture Validator

```bash
# Validate architecture in the current directory using default rules file
axiomod validator architecture

# Validate architecture in a specific directory
axiomod validator architecture ./internal/examples/example

# Validate architecture using a custom rules file
axiomod validator architecture --config ./custom-rules.json
```

### Naming Validator

```bash
# Validate naming conventions in the current directory
axiomod validator naming

# Validate naming conventions in a specific directory
axiomod validator naming ./internal/examples/example

# Specify custom directories for SQL migrations and API handlers
axiomod validator naming --sql ./db/migrations --api ./api/handlers

# Output results in JSON format
axiomod validator naming --json
```

### Domain Validator

```bash
# Validate domain boundaries in the current directory using default rules file
axiomod validator domain

# Validate domain boundaries in a specific directory
axiomod validator domain ./internal/examples/example

# Validate domain boundaries using a custom rules file
axiomod validator domain --config ./custom-rules.json
```

### Static Analysis Validators

```bash
# Run all static analysis tools (vet, staticcheck, gosec)
axiomod validator static-analysis

# Run staticcheck in a specific directory
axiomod validator static-check ./internal/framework/logger

# Run gosec security scanner
axiomod validator security
```

### API Spec Validator

```bash
# Check API spec files in the default directory (e.g., ./api/openapi.yaml)
axiomod validator check-api-spec

# Check API spec files in a specific directory
axiomod validator check-api-spec ./docs/api
```

### Documentation Validator

```bash
# Check if recent code changes have documentation updates
axiomod validator check-docs
```

### Running All Validators

```bash
# Run all validators in the current directory
axiomod validator all

# Run all validators with custom config and directories
axiomod validator all --config ./custom-rules.json --sql ./db/migrations --api ./api/handlers
```
