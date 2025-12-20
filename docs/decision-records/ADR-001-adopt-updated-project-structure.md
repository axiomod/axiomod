
# ADR-001: Adopt Updated Project Structure and Practices for Axiomod Framework

## Status

Accepted

## Context

The Axiomod Framework is intended to be a scalable, enterprise-grade Go framework. Early versions had minor ambiguity regarding configuration management, CLI command organization, package visibility, and example modularity.

To ensure long-term scalability, maintainability, and clarity for contributors and users, several structure and documentation updates were proposed and agreed upon.

## Decision

We adopt the following project structure and practices:

1. **Centralized Configuration Management**
   - All configuration files and code are moved under `internal/framework/config/`.
   - Separate YAML files for different concerns:
     - `cli_config.yaml`
     - `service_default.yaml`
     - `plugin_settings.yaml`
   - Access configurations via helper functions in `config.go` only.

2. **Move Framework Packages**
   - The `pkg/` directory is migrated to `internal/framework/`.
   - All imports updated to reflect internal usage only.

3. **Group CLI Commands**
   - Organize CLI subcommands into logical subdirectories:
     - `core/` (build, deploy, status, logs)
     - `generate/` (handler, service, module generators)
     - `migrate/` (migrations)
     - `plugin/` (plugin management)
     - `validator/` (static analysis, architecture validation)

4. **Example Module Organization**
   - Move `internal/example/` to `internal/examples/example/`.
   - Future example services will also be placed under `internal/examples/`.

5. **Clarify axiomod_server Purpose**
   - If core infrastructure: keep under `cmd/axiomod-server/`.
   - If example application: move under `internal/examples/axiomod_server/`.

6. **Standardize Naming Conventions**
   - Use hyphens (`-`) instead of underscores (`_`) for directories and binaries.
   - Example: rename `axiomod_server` → `axiomod-server`.

7. **Enhance Testing Strategy**
   - Create a `/tests/` directory for unit, integration, and e2e tests.
   - Target minimum 70% code coverage initially.

8. **Documentation Enhancements**
   - Add `docs/decision-records/` to house ADRs.
   - Document how `architecture-rules.json` validators are used in `docs/validator-guide.md`.

## Consequences

- Clear separation of CLI, service, and plugin configuration.
- Stronger enforcement of Go project best practices (internal package safety, domain modularity).
- Improved scalability for future features (multi-example services, plugin expansion).
- Easier onboarding for new contributors.
- Slight short-term refactor effort to migrate packages and imports.
