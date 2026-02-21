# Claude Code Configuration Guide

This guide documents the `.claude/` directory, which configures [Claude Code](https://docs.anthropic.com/en/docs/claude-code) (Anthropic's AI coding assistant) for the Axiomod framework. Contributors use Claude Code to implement features, write tests, review architecture, and maintain documentation while following all project conventions automatically.

## 1. Overview

Claude Code reads project-specific instructions from the `.claude/` directory at the repository root. These instructions teach the AI assistant about Axiomod's architecture, coding standards, plugin system, and development workflows. Every Claude Code session automatically loads the project instructions, ensuring consistent behavior across all contributors.

The configuration consists of:

- **CLAUDE.md** -- Project-wide instructions loaded every session
- **Rules** -- 15 topic-specific rule files that auto-load as context
- **Agents** -- 10 specialized agents for different tasks (architecture review, testing, security scanning, etc.)
- **Skills** -- 5 reusable workflows invoked via slash commands
- **Settings** -- Permission controls for tool access

## 2. Directory Structure

```
.claude/
├── CLAUDE.md                          # Project instructions (loaded every session)
├── settings.json                      # Shared permissions (committed to git)
├── settings.local.json                # Personal permissions (NOT committed)
├── rules/                             # Auto-loaded rule files
│   ├── 01-coding-style.md            # Naming conventions, imports, logging
│   ├── 02-architecture.md            # Layer structure, import rules
│   ├── 03-error-handling.md          # framework/errors usage
│   ├── 04-testing.md                 # Test patterns, libraries, coverage
│   ├── 05-fx-dependency-injection.md # fx module wiring patterns
│   ├── 06-http-delivery.md           # Fiber handler patterns
│   ├── 07-grpc-delivery.md           # gRPC service patterns
│   ├── 08-plugin-system.md           # Plugin interface and lifecycle
│   ├── 09-domain-patterns.md         # Clean Architecture entities, repos, use cases
│   ├── 10-middleware.md              # Middleware struct-with-Handle pattern
│   ├── 11-config-system.md           # Viper config, YAML keys, env vars
│   ├── 12-observability.md           # Logger, tracer, metrics, health
│   ├── 13-resilience.md              # Circuit breaker, retry, fallback
│   ├── 14-database.md               # Connection, transactions, migrations
│   └── 15-ci-build.md               # Build commands, CI pipeline, linting
├── agents/                            # Specialized agent definitions
│   ├── arch-guard.md                 # Read-only architecture validator
│   ├── axiomod-go.md                 # Framework-aware Go developer
│   ├── plugin-builder.md            # Plugin creation specialist
│   ├── domain-scaffolder.md          # Clean Architecture module creator
│   ├── api-designer.md              # Dual-protocol API specialist
│   ├── perf-profiler.md             # Performance analysis specialist
│   ├── unit-tester.md               # Unit test generator
│   ├── security-scanner.md          # Vulnerability scanner
│   ├── requirements-verifier.md     # Post-implementation verifier
│   └── docs-updater.md              # Documentation maintainer
└── skills/                            # Reusable workflow definitions
    ├── validate-arch/SKILL.md        # Architecture validation workflow
    ├── gen-module/SKILL.md           # Domain module scaffolding
    ├── new-plugin/SKILL.md           # Plugin creation workflow
    ├── release/SKILL.md              # Release preparation workflow
    └── add-endpoint/SKILL.md         # Endpoint addition workflow
```

## 3. CLAUDE.md -- Project Instructions

The file `.claude/CLAUDE.md` is the primary project instruction file. Claude Code loads it automatically at the start of every session. It contains:

- **Quick Commands** -- `make build`, `make test`, `make lint`, and all CLI commands
- **Architecture** -- Layer structure and import direction rules
- **fx Module Pattern** -- How to declare and assemble DI modules
- **Plugin System** -- The Plugin interface and registration mechanism
- **Config System** -- Viper-based config with `APP_` env prefix
- **Error Handling** -- `framework/errors` usage and error code mappings
- **Domain Module Structure** -- Clean Architecture 8-package layout
- **Observability** -- Logger, tracer, metrics wrappers
- **Testing Conventions** -- Libraries, patterns, and commands
- **Key Conventions** -- Go version, naming, validation, and prohibitions
- **Post-Implementation Checklist** -- Verification steps after any change

This file acts as the "always-on" context. Every rule, pattern, and convention described here applies to all Claude Code interactions with this repository.

## 4. Rules (`.claude/rules/`)

The `rules/` directory contains 15 Markdown files, each covering a specific topic. Claude Code loads all rule files automatically at the start of every session alongside `CLAUDE.md`. Rules provide detailed, enforceable guidelines with code examples.

### Rule File Reference

| File | Topic | Key Content |
|------|-------|-------------|
| `01-coding-style.md` | Naming and formatting | PascalCase types, snake_case files, import organization (stdlib/internal/third-party), constructor patterns, Go doc comments, structured logging keys |
| `02-architecture.md` | Layer boundaries | Import direction table (entity imports nothing, repo -> entity, etc.), forbidden cross-domain imports, exceptions for test files, 8-package domain structure |
| `03-error-handling.md` | Error patterns | `framework/errors` constructors (`NewNotFound`, `Wrap`, `WithCode`), error codes, HTTP/gRPC mapping, domain errors with `DomainError` struct |
| `04-testing.md` | Test conventions | Table-driven tests with `t.Run()`, `testify/assert`, manual mocks (no gomock), fx integration tests, HTTP handler tests, >80% coverage target |
| `05-fx-dependency-injection.md` | DI wiring | `var Module = fx.Options(...)`, module assembly in `fx_options.go`, interface binding, lifecycle hooks, domain module wiring pattern |
| `06-http-delivery.md` | HTTP handlers | Handler struct with use case dependencies, `RegisterRoutes(fiber.Router)`, handler method pattern (parse -> delegate -> respond), route groups under `/api/v1` |
| `07-grpc-delivery.md` | gRPC services | Service struct embedding `Unimplemented*Server`, method pattern (map request -> execute -> map response), interceptor chain, registration via `fx.Invoke` |
| `08-plugin-system.md` | Plugin interface | 4-method interface (`Name`, `Initialize`, `Start`, `Stop`), lifecycle, built-in plugins (mysql, postgresql, jwt, keycloak, casdoor, casbin), YAML config |
| `09-domain-patterns.md` | Clean Architecture | Entity with `Validate()`, repository interfaces, use case per file with Input/Output, domain services, memory repositories with `sync.RWMutex`, cache and event patterns |
| `10-middleware.md` | Middleware patterns | Struct-with-`Handle()` returning `fiber.Handler`, available middleware table (logging, auth, role, timeout, recovery, metrics, tracing, RBAC), application order |
| `11-config-system.md` | Configuration | `Config` struct hierarchy, YAML camelCase keys, PascalCase Go fields, no struct tags (Viper case-insensitive), `APP_` env prefix, `Provider` interface, specialized loaders |
| `12-observability.md` | Logging/metrics/tracing | `observability.Logger` (zap), `observability.Tracer` (OTel), `observability.Metrics` (Prometheus), pre-defined metric vectors, health check registration |
| `13-resilience.md` | Fault tolerance | Circuit breaker (states, options, thread-safety), resilience wrapper (retry + timeout + fallback), HTTP client with built-in resilience |
| `14-database.md` | Database patterns | `database.Connect()`, `WithTransaction` for auto rollback/commit, query wrappers with metrics and slow query detection, migration CLI commands |
| `15-ci-build.md` | Build and CI | Make targets, ldflags version injection, GitHub Actions pipeline (verify, format, vet, test with race detector, build), CodeQL analysis, pre-submit checklist |

### How Rules Auto-Load

Claude Code discovers and loads all `.md` files in `.claude/rules/` at session start. The files are numbered `01` through `15` for human readability, but Claude Code loads them all regardless of naming. The content becomes part of the assistant's context for the entire session.

## 5. Agents (`.claude/agents/`)

Agents are specialized Claude Code personas, each with a focused role, specific capabilities, and defined restrictions. Contributors invoke agents through natural language. Each agent file contains a YAML front-matter `description` field that defines trigger phrases.

### Agent Reference

#### arch-guard -- Architecture Guard

- **Purpose**: Read-only architecture enforcement. Validates imports against `architecture-rules.json`, checks layer boundaries, and verifies domain module structure.
- **Triggers**: "architecture rules", "import violations", "layer boundaries", "dependency direction"
- **Capabilities**: Reads files, runs `axiomod validator architecture`, reports violations with file:line detail
- **Restrictions**: Never modifies files. Read-only analysis only.

#### axiomod-go -- Axiomod Go Developer

- **Purpose**: Framework-aware Go implementation. Writes handlers, services, middleware, fx modules, tests, and error handling following all Axiomod patterns.
- **Triggers**: "implement", "write", "create", "add feature", "fix bug", "refactor", "Go code"
- **Capabilities**: Full read/write access. Implements features end-to-end with proper fx wiring, error handling, and tests.

#### plugin-builder -- Plugin Builder

- **Purpose**: Creates plugins implementing the Plugin interface. Handles registration, config entries, health checks, and lifecycle methods.
- **Triggers**: "new plugin", "create plugin", "plugin interface", "plugin registry", "extend plugin"
- **Capabilities**: Creates plugin files, registers in `RegisterNewPlugins` or `registerBuiltInPlugins`, adds config entries, writes tests.

#### domain-scaffolder -- Domain Scaffolder

- **Purpose**: Creates and extends Clean Architecture domain modules with the standard 8-package structure.
- **Triggers**: "new domain", "new module", "scaffold", "clean architecture", "add entity", "add use case"
- **Capabilities**: Generates entity, repository, usecase, service, delivery, infrastructure packages, and `module.go` wiring. Can also run `axiomod generate module`.

#### api-designer -- API Designer

- **Purpose**: Designs dual-protocol APIs ensuring HTTP (Fiber) and gRPC endpoints are consistent. Covers routing, interceptor chains, error code mapping, and proto file conventions.
- **Triggers**: "API design", "endpoint", "HTTP route", "gRPC service", "protobuf", "interceptor", "dual protocol"
- **Capabilities**: Designs route structures, handler patterns, gRPC service implementations, and error mapping tables.

#### perf-profiler -- Performance Profiler

- **Purpose**: Diagnoses performance issues using data-driven analysis. Profiles CPU, memory, and latency. Writes benchmarks.
- **Triggers**: "performance", "profiling", "benchmark", "slow", "memory leak", "latency", "throughput", "optimize"
- **Capabilities**: Runs `go test -bench`, `go tool pprof`, `go tool trace`, `go test -race`. Reads metrics and profile output.
- **Restrictions**: Presents findings and recommendations. Does not modify source code directly.

#### unit-tester -- Unit Test Generator

- **Purpose**: Writes thorough, idiomatic unit tests for all layers (entity, repository, use case, handler, middleware, plugin).
- **Triggers**: "write tests", "add tests", "unit test", "test coverage", "create test for"
- **Capabilities**: Creates table-driven tests with `testify`, manual mocks, fx integration tests, Fiber handler tests. Runs `go test -v -race`.

#### security-scanner -- Security Scanner

- **Purpose**: Scans for vulnerabilities across 10 categories: injection, auth, secrets, input validation, crypto, dependencies, concurrency, server config, logging, and path traversal.
- **Triggers**: "security scan", "vulnerability check", "security audit", "check for vulnerabilities", "OWASP", "secret scan"
- **Capabilities**: Reads all files, runs `go vet`, `govulncheck`, `go list -json -m all`. Reports findings with severity levels (CRITICAL/HIGH/MEDIUM/LOW/INFO).
- **Restrictions**: Never modifies files. Never executes exploits. Redacts actual secrets in output.

#### requirements-verifier -- Requirements Verifier

- **Purpose**: Post-implementation quality assurance. Verifies that new code meets requirements, follows conventions, and passes all checks.
- **Triggers**: "verify code", "check requirements", "validate implementation", "review changes", "post-implementation check"
- **Capabilities**: Runs the full verification checklist (architecture, conventions, functionality, DI wiring, tests, build, lint, config, observability). Produces a pass/fail report.
- **Restrictions**: Does not modify files. Only reads and verifies.

#### docs-updater -- Documentation Updater

- **Purpose**: Updates `docs/` to reflect code changes. Identifies which docs are affected by code changes and updates them.
- **Triggers**: "update docs", "update documentation", "sync docs", "docs need update", "document this change"
- **Capabilities**: Reads source code and docs, edits documentation files. Uses `git diff` to identify changes. Follows the change-to-doc mapping table.
- **Restrictions**: Never modifies source code. Never creates release notes unless explicitly asked.

### How to Invoke Agents

Invoke agents by describing your task using the trigger phrases. Claude Code matches the task description to the agent's `description` field and activates the appropriate persona. Examples:

```
"Check for import violations in the new order module"
  -> Activates: arch-guard

"Write a Redis cache plugin"
  -> Activates: plugin-builder

"Add unit tests for the CreateExampleUseCase"
  -> Activates: unit-tester

"Run a security audit on the auth middleware"
  -> Activates: security-scanner
```

You can also explicitly reference an agent by name in your request.

## 6. Skills (`.claude/skills/`)

Skills are reusable, multi-step workflows. Each skill lives in its own subdirectory with a `SKILL.md` file that defines the steps. Skills are invoked via slash commands or natural language triggers defined in the YAML front-matter.

### Skill Reference

#### validate-arch -- Validate Architecture

- **Invocation**: "validate architecture", "check imports", "verify layers", "run arch check"
- **What it does**:
  1. Runs `axiomod validator architecture` against `architecture-rules.json`
  2. Analyzes any violations found (reads violating file, identifies the forbidden import)
  3. Suggests the correct dependency direction and refactoring approach
  4. Performs manual verification for edge cases (cross-domain imports, pattern rules, exceptions)

#### gen-module -- Generate Domain Module

- **Invocation**: "create module", "new domain", "scaffold module", "generate module `<name>`"
- **Inputs**: Module name (required, lowercase, alphanumeric)
- **What it does**:
  1. Validates module name and checks `examples/<name>/` does not exist
  2. Runs `axiomod generate module --name=<name>`
  3. Verifies the generated 8-package structure
  4. Lists post-generation tasks (implement entity fields, wire into `fx_options.go`, run architecture validation)

#### new-plugin -- Create New Plugin

- **Invocation**: "create plugin", "new plugin", "add plugin `<name>`"
- **Inputs**: Plugin name (required, lowercase), config keys (optional)
- **What it does**:
  1. Creates plugin file implementing the 4-method Plugin interface
  2. Creates tests covering `Name()`, `Initialize()`, `Start()`, `Stop()`
  3. Registers the plugin (built-in or external)
  4. Adds config entry to `configs/service_default.yaml`
  5. Verifies with `go build` and `go test`

#### release -- Release Workflow

- **Invocation**: "release", "tag version", "prepare release", "cut release"
- **Inputs**: Version (required, e.g., v1.5.0)
- **What it does**:
  1. Runs pre-release checks (`make deps`, `make build`, `make test`, `make lint`, architecture validation)
  2. Verifies current version in Makefile and git state
  3. Updates VERSION in Makefile
  4. Commits version bump, suggests tag and push (asks for confirmation)
  5. Optionally creates a GitHub release via `gh release create`

#### add-endpoint -- Add Endpoint to Existing Module

- **Invocation**: "add endpoint", "new route", "add API", "new handler method"
- **Inputs**: Module name (required), action name (required, e.g., "update", "delete"), HTTP method/path (optional)
- **What it does**:
  1. Creates a new use case with Input/Output structs
  2. Adds repository method if needed, updates infrastructure implementations
  3. Adds handler method with route registration
  4. Adds gRPC method (default: yes)
  5. Updates `module.go` wiring with `fx.Provide`
  6. Adds tests for the new use case and handler
  7. Verifies with `go build`, `go test`, and `axiomod validator architecture`

## 7. Settings (`.claude/settings.json`)

The `settings.json` file controls which tools Claude Code is permitted to use without asking for confirmation. This file is committed to the repository, so all contributors share the same permissions.

### Allowed Operations

The `allow` list grants automatic permission for:

- **Build tools**: `make`, `go test`, `go build`, `go vet`, `go fmt`, `go mod tidy`, `go mod download`, `golangci-lint`
- **CLI commands**: All `axiomod` subcommands (`generate`, `validator`, `init`, `migrate`, `plugin`, `policy`, `test`, `lint`, `fmt`, `build`, `version`, `healthcheck`)
- **Docker**: `docker build`, `docker compose`
- **Git (read-only)**: `git status`, `git log`, `git diff`, `git branch`, `git tag`
- **GitHub CLI**: `gh pr`, `gh issue`

### Denied Operations

The `deny` list blocks:

- **Sensitive file access**: Reading or editing `.env`, `service_local.yaml`, `*.pem`, `*.key` files
- **Destructive commands**: `rm -rf`, `git push --force`, `git reset --hard`

### Full Settings File

```json
{
  "permissions": {
    "allow": [
      "Bash(make:*)",
      "Bash(go test:*)",
      "Bash(go build:*)",
      "Bash(go vet:*)",
      "Bash(go fmt:*)",
      "Bash(go mod tidy:*)",
      "Bash(go mod download:*)",
      "Bash(golangci-lint:*)",
      "Bash(axiomod generate:*)",
      "Bash(axiomod validator:*)",
      "Bash(axiomod init:*)",
      "Bash(axiomod migrate:*)",
      "Bash(axiomod plugin:*)",
      "Bash(axiomod policy:*)",
      "Bash(axiomod test:*)",
      "Bash(axiomod lint:*)",
      "Bash(axiomod fmt:*)",
      "Bash(axiomod build:*)",
      "Bash(axiomod version:*)",
      "Bash(axiomod healthcheck:*)",
      "Bash(docker build:*)",
      "Bash(docker compose:*)",
      "Bash(git status:*)",
      "Bash(git log:*)",
      "Bash(git diff:*)",
      "Bash(git branch:*)",
      "Bash(git tag:*)",
      "Bash(gh pr:*)",
      "Bash(gh issue:*)"
    ],
    "deny": [
      "Read(*.env)",
      "Read(**/service_local.yaml)",
      "Read(**/*.pem)",
      "Read(**/*.key)",
      "Edit(*.env)",
      "Edit(**/service_local.yaml)",
      "Edit(**/*.pem)",
      "Edit(**/*.key)",
      "Bash(rm -rf:*)",
      "Bash(git push --force:*)",
      "Bash(git reset --hard:*)"
    ]
  }
}
```

Any operation not in the `allow` list prompts the user for confirmation before executing. Any operation in the `deny` list is blocked entirely.

## 8. Local Settings (`.claude/settings.local.json`)

The file `.claude/settings.local.json` holds personal permission overrides. It is **not committed to git** (add it to `.gitignore`). Use it to add permissions specific to your local environment, such as:

```json
{
  "permissions": {
    "allow": [
      "Bash(go run:*)",
      "Bash(curl:*)"
    ]
  }
}
```

Local settings merge with the shared `settings.json`. The `deny` list in the shared settings takes precedence -- you cannot override a denied operation in your local settings.

## 9. Auto-Memory

Claude Code maintains a personal memory file for each contributor at:

```
~/.claude/projects/<project-path-hash>/memory/MEMORY.md
```

This file persists across conversations and stores:

- Project overview and key facts
- Directory map and entry points
- Key dependencies
- Any notes Claude Code records during sessions

Auto-memory is **personal** -- it lives outside the repository and is not shared with other contributors. Each contributor's Claude Code instance builds its own memory over time.

## 10. How to Add New Rules

Rules enforce coding standards and architectural constraints. To add a new rule:

1. **Create a Markdown file** in `.claude/rules/` with the naming convention `<NN>-<topic>.md`, where `<NN>` is the next available number:

   ```bash
   # Example: adding a rule for caching patterns
   touch .claude/rules/16-caching.md
   ```

2. **Write the rule** following the existing format. Include:
   - A top-level heading (`# Topic Name`)
   - Sections with code examples showing correct patterns
   - A `## Rules` section at the end with numbered, enforceable statements

   Example structure:

   ```markdown
   # Caching Patterns

   ## Cache Interface

   All caches implement:

   ```go
   type Cache interface {
       Get(ctx context.Context, key string) (interface{}, error)
       Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
       Delete(ctx context.Context, key string) error
   }
   ```

   ## Key Naming

   Use the pattern `<entity>:<id>`:

   ```go
   key := fmt.Sprintf("example:%s", id)
   ```

   ## Rules

   1. Always set a TTL on cache entries
   2. Use `<entity>:<id>` key format
   3. Handle cache misses gracefully -- fall back to the repository
   ```

3. **Commit the file** -- it takes effect for all contributors on the next Claude Code session.

### Guidelines for Writing Rules

- Keep each rule file focused on a single topic
- Include both correct and incorrect code examples where helpful
- Reference actual project paths and package names
- End with a numbered `## Rules` section for quick reference
- Aim for 50-150 lines per rule file to keep context manageable

## 11. How to Add New Agents

Agents provide specialized personas for specific tasks. To add a new agent:

1. **Create a Markdown file** in `.claude/agents/` named `<agent-name>.md`:

   ```bash
   touch .claude/agents/migration-helper.md
   ```

2. **Add YAML front-matter** with a `description` field. This field defines when the agent activates and what trigger phrases invoke it:

   ```markdown
   ---
   description: "Database migration specialist. Invoke when creating, running, or troubleshooting migrations. Triggers on 'migration', 'migrate', 'schema change', 'database migration'."
   ---
   ```

3. **Write the agent body** with:
   - A role statement ("You are a...")
   - Responsibilities (numbered list of what the agent does)
   - Patterns and code examples relevant to the agent's domain
   - Tool restrictions (what the agent may and may not do)

   Example structure:

   ```markdown
   ---
   description: "Database migration specialist. Triggers on 'migration', 'migrate', 'schema change'."
   ---

   # Migration Helper Agent

   You are a database migration specialist for the Axiomod framework.

   ## Your Responsibilities

   1. Create migration files using `axiomod migrate create <name>`
   2. Verify migrations apply cleanly with `axiomod migrate up`
   3. Validate rollback with `axiomod migrate down`

   ## Migration File Pattern

   ```sql
   -- +migrate Up
   CREATE TABLE foo (...);

   -- +migrate Down
   DROP TABLE IF EXISTS foo;
   ```

   ## Tool Restrictions

   - You may run `axiomod migrate` commands
   - You may read and edit migration files in `migrations/`
   - You must NOT modify application source code
   ```

4. **Commit the file** -- it becomes available to all contributors immediately.

### Guidelines for Writing Agents

- Give each agent a clear, bounded responsibility
- Include trigger phrases in the `description` front-matter that match natural language requests
- Define explicit tool restrictions (what the agent may and may not do)
- Reference specific project paths, commands, and patterns
- Keep the agent focused -- prefer creating multiple specialized agents over one broad agent

## 12. Token Budget

All `.claude/` configuration files auto-load at the start of every Claude Code session. Understanding the token cost helps keep the configuration efficient.

### Current Budget Breakdown

| Category | Files | Approximate Tokens |
|----------|-------|--------------------|
| `CLAUDE.md` | 1 | ~1,800 |
| Rules (`rules/`) | 15 | ~9,500 |
| Agents (`agents/`) | 10 | ~6,500 |
| Skills (`skills/`) | 5 | ~1,700 |
| Settings (`settings.json`) | 1 | ~200 |
| **Total** | **32 files** | **~19,700** |

This represents approximately **1.4-2% of Claude Code's available context window** (~200K tokens). The remaining context is available for source code, tool outputs, and conversation history.

### Keeping the Budget Manageable

- **Rules**: Each rule file averages ~630 tokens. Keep individual rules under 150 lines.
- **Agents**: Each agent file averages ~650 tokens. Focus on patterns and restrictions, not exhaustive documentation.
- **Skills**: Each skill file averages ~340 tokens. Skills are procedural step lists, kept concise by design.
- **Avoid duplication**: If the same information appears in `CLAUDE.md` and a rule file, keep the summary in `CLAUDE.md` and the details in the rule.
- **Measure impact**: Adding a new rule of typical length increases session overhead by ~0.3%. Adding a new agent increases it by ~0.3%.

### When to Add vs. When to Edit

- Add a new rule file when a genuinely new topic needs enforcement (e.g., a new subsystem or pattern).
- Edit an existing rule file when the topic already exists but needs updates.
- Avoid creating rules for niche topics that apply to only one file or function -- use inline code comments instead.
