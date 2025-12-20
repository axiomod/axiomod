
# Implementation Plan: `axiomod` CLI + `Makefile` for Go Macroservice Framework

---

# ✨ Overview
This document presents a **complete, detailed, and technical implementation plan** for:
- **`axiomod` Cobra CLI** tool: External user-facing CLI for project generation, module management, migration, etc.
- **`Makefile`**: Internal developer automation for building, testing, linting, and running services.

Both tools are organized within the Go Macroservice Framework structure.

---

# 🏗 Project Structure (Updated)

```plaintext
/go-macroservice-framework
├── cmd/
│   ├── macroservice/                
│   │   ├── main.go
│   │   ├── wire.go
│   │   └── fx_options.go
│   └── axiomod/                   
│       ├── main.go
│       ├── root.go
│       └── commands/
│           ├── init.go
│           ├── generate/
│           │   ├── module.go
│           │   ├── service.go
│           │   └── handler.go
│           ├── migrate/
│           │   ├── create.go
│           │   ├── up.go
│           │   └── down.go
│           ├── config/
│           │   ├── validate.go
│           │   └── diff.go
│           ├── test.go
│           ├── lint.go
│           ├── fmt.go
│           ├── build.go
│           ├── dockerize.go
│           ├── deploy.go
│           ├── status.go
│           ├── logs.go
│           ├── healthcheck.go
│           ├── plugin/
│           │   ├── install.go
│           │   ├── list.go
│           │   └── remove.go
│           ├── interactive.go
│           └── version.go
├── internal/
├── pkg/
├── scripts/
│   ├── docker-compose.yml
│   └── Makefile
├── api/
├── docs/
└── README.md
```

---

# 🔥 `axiomod` CLI Features

- `init [project-name]`
- `generate module [name]`
- `generate service [name]`
- `generate handler [name]`
- `migrate create [name]`
- `migrate up`
- `migrate down`
- `config validate`
- `config diff [env1] [env2]`
- `test`
- `lint`
- `fmt`
- `build`
- `dockerize`
- `deploy [env]`
- `status`
- `logs [service]`
- `healthcheck`
- `plugin install [plugin]`
- `plugin list`
- `plugin remove [plugin]`
- `interactive`
- `version`

---

# 🔥 Makefile Features

- `build`
- `build-cli`
- `build-all`
- `run`
- `test`
- `test-cover`
- `lint`
- `fmt`
- `vet`
- `deps`
- `update-deps`
- `generate`
- `proto`
- `mock`
- `docker-build`
- `docker-run`
- `docker-push`
- `migrate`
- `migrate-new`
- `migrate-rollback`
- `clean`
- `reset`
- `help`

---

# 🛠 Detailed Task List

## Phase 1: CLI Setup
- [ ] Scaffold `/cmd/axiomod/`
- [ ] Initialize Cobra CLI
- [ ] Create `root.go` and basic command structure

## Phase 2: Implement Core CLI Commands
- [ ] Implement `init` command
- [ ] Implement `generate module`, `generate service`, `generate handler`
- [ ] Implement `migrate create`, `migrate up`, `migrate down`
- [ ] Implement `config validate`, `config diff`
- [ ] Implement `test`, `lint`, `fmt`
- [ ] Implement `build`, `dockerize`, `deploy`
- [ ] Implement `status`, `logs`, `healthcheck`
- [ ] Implement plugin system: `plugin install/list/remove`
- [ ] Implement `interactive` command
- [ ] Implement `version` command

## Phase 3: Makefile Setup
- [ ] Create `/scripts/Makefile`
- [ ] Define targets for build/run/test
- [ ] Define targets for lint/format/deps
- [ ] Define targets for docker operations
- [ ] Define targets for migrations
- [ ] Add `help` documentation

## Phase 4: Testing & Validation
- [ ] Write unit tests for CLI commands
- [ ] Test Makefile commands locally
- [ ] Integrate into CI/CD pipeline

---

# ⚙️ Technical Stack & Libraries

| Component           | Technology            |
|:-------------------|:-----------------------|
| CLI Framework       | Cobra (spf13/cobra)     |
| Config Management   | Viper (spf13/viper)     |
| Proto Compilation   | protoc + plugins       |
| Linting             | golangci-lint          |
| Mock Generation     | mockgen (golang/mock)   |
| Docker Integration  | Docker CLI             |
| Migrations          | Custom via `axiomod`  |

---

# 🚀 Final Note

This plan ensures a powerful, scalable, and developer-friendly CLI + automation system for the Go Macroservice Framework. 
