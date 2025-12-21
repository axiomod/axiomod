# Release Notes v1.4.0

## What's New

### Deep Health Checks (Observability)

- Introduced deep health checks for Database and Kafka connectivity.
- Enhanced `/ready` endpoint to provide granular component status.
- Framework now supports registering custom health checks via `health.Health` instance.
- Plugins can now register their own health checks during initialization.

### Dynamic Configuration Reloading

- Implemented file watching for configuration changes.
- Added `WatchConfig` capability to the configuration provider.
- Changes in the configuration file now trigger notifications within the application.

### Plugin Ecosystem

- Added unit tests for all core plugin stubs (LDAP, SAML, Multi-Tenancy, Audit, ELK).
- Refactored Plugin interface to include `health.Health` dependency for self-monitoring.
- Improved plugin lifecycle management.

### Dependency Management

- Audited all project dependencies.
- Added comprehensive documentation in `docs/dependencies.md`.
- Cleaned and verified `go.mod`.

### Infrastructure

- Added detailed GitHub Actions CI pipeline (`.github/workflows/ci.yml`) including:
  - Format checks
  - Static analysis (`go vet`)
  - Race condition detection
  - Build verification

## Upgrading

No breaking changes in the API. Ensure your `service_default.yaml` is up to date if you wish to use the new health check features, though they utilize existing connection parameters.
