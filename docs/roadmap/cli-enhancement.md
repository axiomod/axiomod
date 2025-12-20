# Axiomod CLI: Enterprise Enhancement Roadmap

This document outlines the strategic enhancements required to transform the `axiomod` CLI into a comprehensive enterprise governance and productivity tool.

## 1. Feature Scaffolding 2.0 (Domain-First)

**Current State**: Generates isolated files (handler, service, module).
**Enterprise Need**: Generate complete, compliant vertical slices from a specification.

- **OpenAPI-Driven Generation**:
  - Command: `axiomod generate from-spec api.yaml`
  - Feature: Automatically generate DTOs, Handlers, and Routes from an OpenAPI 3.0 definition.
  - Benefit: Ensures implementation matches the contract 100%.

- **Domain-Driven Scaffolding**:
  - Command: `axiomod generate domain Order --aggregate`
  - Feature: Generates the full clean architecture tree: `internal/domain/order` (Entities, Repository Interfaces), `internal/usecase/order`, and `internal/delivery/http/order`.

## 2. Advanced Governance & Compliance (Policy-as-Code)

**Current State**: Hardcoded logic validation checks.
**Enterprise Need**: Customizable, centralized rule enforcement without recompiling the CLI.

- **Custom Rule Engine**:
  - Feature: Load rules from `.axiomod/rules.yaml` or `.rego` checks.
  - Example: "No packages allowed to import `database/sql` directly except `repository` layer."
- **CI/CD Gatekeeper Mode**:
  - Command: `axiomod gatekeeper run --strict`
  - Feature: A specialized mode for CI pipelines that outputs JUnit/SARIF detailed reports and fails the build on any critical violation.

## 3. Monorepo & Workspace Support

**Current State**: `init` assumes a single-module creation.
**Enterprise Need**: Managing multiple services within a single repository (Google/Meta style).

- **Workspace Management**:
  - Command: `axiomod workspace init`
  - Command: `axiomod app new billing-service` (creates `apps/billing-service`)
  - Feature: Shared `go.mod` management and shared `pkg/` libraries for all apps in the workspace.

## 4. DevOps & Cloud Native Integration

**Current State**: Basic Dockerfile generation.
**Enterprise Need**: Full K8s/Helm and Observability setup.

- **Infrastructure-as-Code (IaC) Generator**:
  - Command: `axiomod infra generate --provider=aws --type=terraform`
  - Feature: Generates Terraform/Pulumi scripts for deploying the service with RDS, ElastiCache, and IAM roles.
- **Helm Chart Scaffolding**:
  - Command: `axiomod deploy generate-chart`
  - Feature: Creates a production-ready Helm chart with sidecars (Envoy/Istio) and `ServiceMonitor` for Prometheus.

## 5. Extensible Plugin Architecture

**Current State**: Internal plugin logic.
**Enterprise Need**: Allow third-party teams to extend the CLI.

- **Binary Plugin System**:
  - Structure: Look for binaries named `axiomod-foo` in `$PATH`.
  - Command: `axiomod foo` delegates to the binary.
  - Benefit: Platform teams can release their own `axiomod-company-standards` plugin without forking the core CLI.

## 6. Database Migration "Time Travel"

**Current State**: Basic `up`/`down`.
**Enterprise Need**: Safety nets for production data.

- **Dry-Run & SQL Preview**:
  - Command: `axiomod migrate up --dry-run`
  - Feature: Prints the exact SQL that will be executed without running it.
- **Rollback Testing**:
  - Command: `axiomod migrate test`
  - Feature: Spins up a Docker container, runs `up`, inserts data, runs `down`, and verifies integrity.

## Summary of New Commands

| Category | New Command | Description |
| :--- | :--- | :--- |
| **Scaffold** | `axiomod generate from-spec` | Generate code from OpenAPI |
| **Scaffold** | `axiomod workspace init` | Initialize Monorepo structure |
| **Gov** | `axiomod gatekeeper` | CI/CD strict compliance runner |
| **Ops** | `axiomod infra generate` | Generate Terraform/Helm charts |
| **Db** | `axiomod migrate test` | Verify migration rollback safety |
