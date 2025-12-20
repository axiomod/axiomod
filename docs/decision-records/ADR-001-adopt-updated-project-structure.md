
# ADR-001: Adopt Public Project Structure for Axiomod Library Readiness

## Status

Accepted (Revised)

## Context

The Axiomod Framework was initially designed with core components in an `internal/` package structure. While this provided good encapsulation during early development, it created significant friction for users who wanted to import the framework as a library or use the CLI to bootstrap new projects. Enterprise readiness requires that the framework be easily importable and that generated projects can reference framework types and utilities directly.

## Decision

We adopt a public, library-ready project structure where the core framework components are moved from `internal/` to top-level, public directories:

1. **Top-Level Public Packages**
   - `framework` is moved to `framework/`
   - `platform` is moved to `platform/`
   - `plugins` is moved to `plugins/`
   - All import paths are standardized to `github.com/axiomod/axiomod/{package}`.

2. **Unified Configuration System**
   - Configuration management remains centralized in `framework/config/`.
   - The system supports loading from `framework/config/`, `./config/`, and environment variables.
   - CLI tools use `framework/config/` for default behavior.

3. **CLI Scaffolding Update**
   - Generated projects now import framework components from `github.com/axiomod/axiomod/...`.
   - The CLI handles path transformations and keeps the generated `go.mod` clean.
   - A `replace` directive is provided in examples for local development.

4. **Architecture Validation**
   - `architecture-rules.json` is updated to enforce dependencies between the new public packages.
   - The validator now scans the entire repository root instead of just `internal/`.

5. **Renaming for Clarity**
   - All references use a consistent naming scheme: `axiomod` for the framework and `axiomod` for the CLI tool.
   - Project binaries use hyphenated names (e.g., `axiomod-server`).

## Consequences

- **Importability:** The framework can now be used as a standard Go library in external projects.
- **CLI Robustness:** Scaffolding now produces functional code out-of-the-box.
- **Maintainability:** Clear separation between `framework/` (core logic), `platform/` (runtime/infra), and `plugins/` (extensibility).
- **Ecosystem Ready:** The library structure follows standard Go community conventions for open-source frameworks.
- **One-Time Refactor:** This change required a significant update to all import paths and internal CLI logic, which is now complete.
