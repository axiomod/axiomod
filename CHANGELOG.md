# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Ent as the framework's default ORM (`platform/ent` schema + generated
  client, `NewClientFromDB`), selectable via the new `database.orm` config
  key (`ent` | `sql`).
- Real MySQL support: driver registration in `framework/database`,
  driver-aware DSNs, migrate CLI support with database bootstrap.
- Example module registered in the server with full CRUD over HTTP and gRPC
  (update/delete/list use cases, domain-error → status-code mapping) and a
  full-stack integration test.
- `axiomod init --dev/--framework-path`; generated projects compile out of
  the box (pinned go.mod, gofmt-clean templates, Dockerfile).
- Default Casbin model/policy files (`configs/rbac_model.conf`,
  `configs/rbac_policy.csv`).
- `make validate-arch`, CI architecture-validation and scaffold smoke-test
  steps; GoReleaser config + tag-triggered release workflow.
- Kubernetes manifests under `deploy/kubernetes/`.
- Real `validator check-api-spec` (OpenAPI 3.x via kin-openapi); working
  `status`, `config validate`, `config diff`; `test --unit/--integration`;
  `dockerize --tag`; honest `deploy --dry-run`.
- CONTRIBUTING, SECURITY, CODE_OF_CONDUCT, CHANGELOG.

### Changed

- `framework/observability` moved from `platform/observability`: the
  dependency direction now matches the documented layering, and the
  architecture validator passes with 0 violations (was 94).
- Canonical config is `configs/service_default.yaml`; the loader searches
  `configs/`, code-level defaults restored, Docker entrypoint passes
  `-config` and the image has a HEALTHCHECK.
- Metrics are served at `/metrics` on the HTTP port; the unused
  `observability.metricsPort` knob was removed.
- `server.Module` provides `*fiber.App` and `grpc.Module` provides
  `*grpc.Server` so the documented domain-wiring pattern resolves.
- OIDC is opt-in: idle when `auth.oidc.issuerUrl` is empty.

### Removed

- Committed binaries at the repository root.
- Stale divergent `framework/config/service_default.yaml`.
- `logs` CLI stub, unused `framework/router` package, `scripts/run_tests.go`,
  the example module's format-only auth middleware.

## [0.2.0] - 2025-12-30

Claude Code integration: `.claude/` configuration (15 rules, 10 agents,
5 skills), contributor guide. See
[docs/release-notes/v0.2.0.md](docs/release-notes/v0.2.0.md).

## [0.1.0] - 2025-12-21

Initial release: Clean Architecture + fx DI core, dual-protocol serving
(Fiber HTTP + gRPC), observability stack, plugin system, CLI. See
[docs/release-notes/v0.1.0.md](docs/release-notes/v0.1.0.md).
