# Axiomod Launch-Readiness Task Plan

**Audience:** AI coding agents (Claude Code or similar) executing tasks autonomously.
**Source:** [`feature-implementation-audit.md`](./feature-implementation-audit.md) (audit-id `2026-06-12-launch-readiness`). Each task references the audit finding it resolves.
**Goal:** Close every P0/P1 gap between documentation and implementation before public launch.

---

## 0. Agent operating rules (read before any task)

1. **One task per branch/PR.** Branch naming: `fix/AX-NNN-short-slug`. Reference the task ID in the commit message.
2. **Definition of done — every task must pass ALL of these before it is complete:**
   ```bash
   gofmt -l .                      # must print nothing
   go vet ./...
   go build ./...
   go test -race ./...
   golangci-lint run ./...         # via `make lint` (installs pinned version via `make deps`)
   ```
   After AX-012 lands, additionally: `./bin/axiomod validator architecture` must exit 0.
3. **Follow `.claude/CLAUDE.md` and `.claude/rules/*`:** `framework/errors` for error wrapping (no raw `fmt.Errorf` in framework code), no `fmt.Println` in library code (CLI command output via `fmt` is acceptable — that is the existing CLI idiom), table-driven tests with testify, every exported symbol gets a doc comment, new fx modules registered in `cmd/axiomod-server/fx_options.go`.
4. **Docs move with code.** If a task changes behavior, update every doc that states the old behavior in the same PR (`grep -rn` the old value across `README.md`, `docs/`, `.claude/`).
5. **Do not bump the version, create tags, or rewrite git history** — those are maintainer actions (see Decisions D3 and AX-031).
6. **No scope creep.** Each task lists "Out of scope" pointers where adjacent temptations exist.
7. When a task says **VERIFY EMPIRICALLY**, run the binary/command and paste the observed output into the PR description.

---

## 1. Maintainer decisions (defaults pre-selected so agents are not blocked)

| ID | Decision | Recommended default (agents proceed with this unless overridden) |
|---|---|---|
| **D1** | Resolve the layering contradiction: `framework/*` imports `platform/observability`, but docs say platform sits above framework. | **Move `platform/observability` → `framework/observability`** (option A). Pre-release with zero tags, an import-path break is acceptable and makes the documented rule ("platform may import framework, never the reverse") true. |
| **D2** | MySQL: advertised but no driver exists. | **Implement it** (driver + migrate support + tests). It is a small task and removal would shrink the launch story. |
| **D3** | ~62 MB of macOS binaries exist in git history. | Agent removes them from the working tree now (AX-007). **History purge (`git filter-repo`) is a maintainer-only, destructive step** scheduled before the repo is publicized. |
| **D4** | Metrics port is 9100 in docs/CLAUDE.md/compose but 9091 in `configs/service_default.yaml`/Dockerfile. | **Standardize on 9100**, after AX-004 first determines whether `metricsPort` even drives a listener. |

---

## 2. Execution waves

Waves are ordered to minimize rebase pain (the Wave 2 refactor touches ~30 files — land it before the many small Wave 3+ PRs). Tasks **within** a wave are independent and parallelizable unless `Depends` says otherwise.

| Wave | Theme | Tasks |
|---|---|---|
| 1 | Foundation correctness | AX-001 … AX-008 |
| 2 | Architecture truth | AX-010 … AX-012 |
| 3 | First-run experience | AX-020 … AX-026 |
| 4 | Release engineering | AX-030 … AX-031 |
| 5 | CLI integrity | AX-040 … AX-048 |
| 6 | Docs reconciliation | AX-050 … AX-057 |
| 7 | Quality bar | AX-060 … AX-066 |
| 8 | Post-launch (P2) | AX-070 … AX-074 |

---

## WAVE 1 — Foundation correctness (P0)

### AX-001 · Config loader: search `configs/` and make one file canonical
**Priority:** P0 · **Size:** S · **Depends:** —
**Problem:** The documented default config is `configs/service_default.yaml`, but `framework/config/config.go:76-90` searches `.`, `config/`, `../config/`, `../../config/`, `framework/config/`, `/etc/app` — never `configs/`. A second, divergent `framework/config/service_default.yaml` (gRPC 50051, missing observability/database sections) is what actually loads. Empirically proven: the server boots gRPC on :50051 instead of the configured :9090.
**Files:** `framework/config/config.go`, `framework/config/service_default.yaml` (delete), `configs/service_default.yaml`, `framework/config/config_test.go`.
**Steps:**
1. Add `configs`, `../configs`, `../../configs` to the Viper search paths, **before** `framework/config` in priority order.
2. Delete `framework/config/service_default.yaml`. Search for anything reading it directly (`grep -rn "framework/config/service_default"`) and update (note: `cmd/axiomod/cmd/core/config/validate.go` scans `framework/config/*.yaml` — point it at `configs/`; `init.go` next-steps text mentions it — fixed properly in AX-021 but adjust if touched here).
3. Ensure `configs/service_default.yaml` is complete and becomes the single source of truth (it already has all sections).
4. Add a table-driven test: temp dir with `configs/service_default.yaml`, chdir, `config.Load("")` → asserts values from that file (e.g. `GRPC.Port == 9090`).
**Acceptance:** From repo root, `go run ./cmd/axiomod-server` logs gRPC on `0.0.0.0:9090` and service name `axiomod-service`. VERIFY EMPIRICALLY (boot ≤10 s, capture log).
**Out of scope:** Docker entrypoint (AX-003), code defaults (AX-002).

### AX-002 · Config: restore code-level defaults
**Priority:** P0 · **Size:** S · **Depends:** AX-001
**Problem:** `framework/config/config.go:300-302` has `SetDefault` calls commented out. If no config file is found (e.g. today's Docker image), the app runs with zero values — HTTP/gRPC ports 0, no app name.
**Steps:** Implement a `setDefaults(v *viper.Viper)` called in the loader with sane defaults matching `configs/service_default.yaml`: `app.name=axiomod-service`, `app.environment=development`, `http.host=0.0.0.0`, `http.port=8080`, `grpc.host=0.0.0.0`, `grpc.port=9090`, `observability.logLevel=info`, `observability.logFormat=json`, `observability.metricsEnabled=true`, plus database pool defaults already used in `framework/database`. Add a test: `config.Load("")` in an empty temp dir returns those defaults, no error.
**Acceptance:** Running the server binary from an empty directory boots HTTP :8080 / gRPC :9090 instead of failing or binding random ports.

### AX-003 · Docker image actually loads its config
**Priority:** P0 · **Size:** S · **Depends:** AX-001
**Problem:** `Dockerfile` copies `configs/` to `/app/configs/` but `ENTRYPOINT ["/app/axiomod-server"]` passes no `-config`, and (pre-AX-001) the loader never searched `configs/`. The shipped container ran configless; `EXPOSE 8080 9090 9091` advertised ports the process didn't use.
**Steps:**
1. After AX-001 the relative `configs/` path resolves from `WORKDIR /app` — still make it explicit: `ENTRYPOINT ["/app/axiomod-server", "-config", "/app/configs/service_default.yaml"]`.
2. Reconcile `EXPOSE` with final ports (8080, 9090, and the D4 metrics port after AX-004).
3. Add `HEALTHCHECK CMD wget -qO- http://127.0.0.1:8080/live || exit 1` (alpine has wget; add curl/wget if the runtime image lacks it).
**Acceptance:** `docker build -t axiomod-test . && docker run --rm axiomod-test` (or, if no Docker in CI sandbox: replicate by running the binary with that exact `-config` path from a scratch dir) shows gRPC :9090 and config-derived app name. Update `docker-compose.reference.yaml` if ports change.

### AX-004 · Metrics port: investigate and standardize (D4)
**Priority:** P0 · **Size:** S · **Depends:** —
**Problem:** `metricsPort` is 9100 in docs/`.claude`/compose, 9091 in `configs/service_default.yaml`/Dockerfile. Additionally `/metrics` is served on the main Fiber app (port 8080) via `platform/server` — it is unclear whether `metricsPort` starts any listener at all.
**Steps:**
1. Trace `ObservabilityConfig.MetricsPort` usage (`grep -rn MetricsPort`). If it drives no listener: either (a) implement a dedicated metrics listener on that port in `platform/observability`/`platform/server`, or (b) delete the knob and document that metrics live at `:8080/metrics`. Prefer (b) for simplicity unless a separate port is trivially wired.
2. Whichever outcome: make `configs/service_default.yaml`, `Dockerfile EXPOSE`, `docker-compose.reference.yaml`, `configs/prometheus.yml`, `docs/observability-guide.md`, `docs/deployment-guide.md`, `.claude/CLAUDE.md`, `.claude/rules/11-config-system.md`, `.claude/rules/15-ci-build.md` agree on one value (9100 if a port survives).
**Acceptance:** `grep -rn "9091\|9100" --include='*.md' --include='*.yaml' --include='*.yml' .` shows a single consistent story; `curl localhost:<port>/metrics` (or :8080/metrics) returns Prometheus text. VERIFY EMPIRICALLY.

### AX-005 · MySQL support: make it real (D2)
**Priority:** P0 · **Size:** M · **Depends:** —
**Problem:** README/docs advertise a MySQL plugin, but `go.mod` has no MySQL driver — `sql.Open("mysql", …)` at `framework/database/database.go:85` fails with `unknown driver`. The migrate CLI (`cmd/axiomod/cmd/migrate/utils.go`) is also hardcoded postgres-only.
**Steps:**
1. `go get github.com/go-sql-driver/mysql` and blank-import it where `lib/pq` is registered (locate with `grep -rn '"github.com/lib/pq"'` — keep driver registration in one place, likely `framework/database`).
2. Verify/extend `buildDSN()` in `framework/database/database.go` to emit a correct MySQL DSN (`user:pass@tcp(host:port)/db?parseTime=true`) and keep password redaction working for both formats.
3. `cmd/axiomod/cmd/migrate/utils.go`: add a `mysql` branch to `getDSN()` and import `github.com/golang-migrate/migrate/v4/database/mysql`; keep the explicit error for genuinely unsupported drivers. Note `ensureDatabaseExists()` uses postgres `template1` — implement the MySQL equivalent (`CREATE DATABASE IF NOT EXISTS`).
4. Tests: unit-test DSN building + redaction for both drivers (no live DB needed). Gate any live-connection test behind `RUN_INTEGRATION_TESTS=true` per `.claude/rules/04-testing.md`.
5. Update `docs/database-guide.md` to state real driver support (and finish its truncated §4 — coordinate with AX-054).
**Acceptance:** With driver `mysql` configured, `database.Connect` reaches the dial stage (connection-refused error, not `unknown driver`) — assert exactly that in a test.

### AX-006 · Ship default Casbin model + policy files
**Priority:** P0 · **Size:** S · **Depends:** —
**Problem:** `configs/service_default.yaml` references `rbac_model.conf` / `rbac_policy.csv` that don't exist anywhere in the repo, so the casbin plugin, RBAC middleware, and all `axiomod policy` commands fail out of the box (audit §3, B7).
**Steps:**
1. Add `configs/rbac_model.conf` — the standard Casbin RBAC model (`[request_definition] r = sub, obj, act`, role grouping `g`, matcher `g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act`).
2. Add `configs/rbac_policy.csv` with a minimal commented example (`p, role:admin, /api/v1/*, *` and one `g` line).
3. Confirm the paths in `configs/service_default.yaml` resolve relative to repo root; fix paths if needed.
4. Test: `auth.NewRBACService(cfg.Casbin)` constructs successfully with the shipped files; `Enforce` returns expected allow/deny for the sample policy.
**Acceptance:** `./bin/axiomod policy list` exits 0 from a clean checkout and prints the sample policy. VERIFY EMPIRICALLY.

### AX-007 · Remove committed binaries from the working tree
**Priority:** P0 · **Size:** S · **Depends:** —
**Problem:** Two macOS Mach-O binaries (`axiomod` ~29 MB, `axiomod-server` ~33 MB) are tracked at the repo root of a Linux-targeted framework (audit B4).
**Steps:** `git rm axiomod axiomod-server`; add `/axiomod` and `/axiomod-server` (root-anchored, so `cmd/axiomod/` is unaffected) to `.gitignore`; note in the PR that history purge is maintainer decision D3.
**Acceptance:** `git ls-files | grep -E '^axiomod(-server)?$'` is empty; build still works (`make build build-cli` outputs to `bin/`).

### AX-008 · Fix mislabeled release note
**Priority:** P0 · **Size:** XS · **Depends:** —
**Problem:** `docs/release-notes/v0.1.0.md` is internally titled "v1.0.0 - Enterprise Ready" (audit B4).
**Steps:** Retitle to `v0.1.0 - Initial Release`, soften "Enterprise Ready" claims to match the actual v0.1.0 scope. Check `docs/README.md` link text consistency.
**Acceptance:** No occurrence of `v1.0.0` in `docs/release-notes/`.

---

## WAVE 2 — Architecture truth (P0)

### AX-010 · Move `platform/observability` → `framework/observability` (D1)
**Priority:** P0 · **Size:** L · **Depends:** Wave 1 complete (avoid conflicts)
**Problem:** Docs/CLAUDE.md define import direction "platform → framework only", but 7+ framework packages (grpc, database, auth, middleware, kafka, router, worker) import `platform/observability` — the documented layering is the inverse of reality and is the root cause of most of the validator's 94 self-violations (audit B3).
**Steps:**
1. `git mv platform/observability framework/observability`; update the package's own references if any path-derived strings exist.
2. Mechanical import rewrite repo-wide: `github.com/axiomod/axiomod/platform/observability` → `github.com/axiomod/axiomod/framework/observability` (touches framework/*, platform/server, plugins/*, cmd/*, examples/*, tests/*, and the `init.go` + `generate` templates in `cmd/axiomod/cmd/`).
3. Update docs and rules that name the old path: `.claude/CLAUDE.md`, `.claude/rules/02-architecture.md`, `.claude/rules/12-observability.md`, `docs/architecture.md`, `docs/developer-guide.md`, `docs/observability-guide.md`, `docs/technical/framework-core.md`, `architecture-rules.json`.
4. Keep `platform/server` as-is (it legitimately imports framework).
**Acceptance:** Full DoD gate passes; `grep -rn "platform/observability" --include='*.go' .` returns nothing; server boots. VERIFY EMPIRICALLY.
**Out of scope:** rules-file rewrite beyond the path rename (AX-011).

### AX-011 · Rewrite `architecture-rules.json` to encode the real, documented rules
**Priority:** P0 · **Size:** M · **Depends:** AX-010
**Problem:** The validator flags imports the docs explicitly allow (e.g. `plugins/plugin.go:10: plugins imports platform/observability (not allowed)`), and `cmd/*` (18 violations) has no sensible rule. Rules also reference nonexistent `platform/ent/{schema,migrate}`.
**Steps:**
1. Encode: `framework/*` → framework-internal only; `platform/*` → `framework/*`; `plugins/*` → `framework/*` + `platform/*`; `cmd/*` → anything (entry points); `examples/<domain>` internal layer rules unchanged; cross-domain still forbidden.
2. Keep the `platform/ent/{schema,migrate}` exceptions and the "repositories may import platform/ent" pattern rule — `platform/ent` becomes real in AX-026.
3. Keep `_test.go`, `mock_`, `testdata` exceptions.
**Acceptance:** `./bin/axiomod validator architecture` exits 0 on the repo. VERIFY EMPIRICALLY (paste the summary: 0 violations).

### AX-012 · Enforce the validator in CI and Makefile
**Priority:** P0 · **Size:** S · **Depends:** AX-011
**Steps:** Add `validate-arch` target to `Makefile` (build CLI → run `bin/axiomod validator architecture`); add a CI step in `.github/workflows/ci.yml` after build; update `.claude/rules/15-ci-build.md` and `docs/validator-guide.md` CI examples (they currently show GitLab CI — add the real GitHub Actions step).
**Acceptance:** CI workflow file contains the step; `make validate-arch` exits 0 locally.

---

## WAVE 3 — First-run experience (P0)

### AX-020 · `axiomod init`: generate a project that compiles
**Priority:** P0 · **Size:** M · **Depends:** AX-010 (templates import new paths)
**Problem (audit B1):** Generated `go.mod` contains only `replace github.com/axiomod/axiomod => ../` — valid only inside the framework checkout — and pins nothing, so `go mod tidy` fails on `cloud.google.com/go/compute/metadata: ambiguous import`.
**Steps (in `cmd/axiomod/cmd/core/init.go`):**
1. Generate `require github.com/axiomod/axiomod <version>` where `<version>` comes from `framework/version.Version` (ldflags-injected); fall back to `@latest` semantics only if version is `dev`.
2. Drop the `replace` by default. Add `--dev` flag that emits `replace github.com/axiomod/axiomod => <abs-or-rel path>` resolved from an explicit `--framework-path` value or by walking up from CWD to find a `go.mod` declaring module `github.com/axiomod/axiomod` (error clearly if not found).
3. Also pin `go.uber.org/fx` and `go.uber.org/zap` requires (versions matching the framework's `go.mod` — read them at build time or hardcode and assert in a test against `go.mod`).
4. Fix the post-init next-steps text: it says "Update framework/config/service_default.yaml" but generates `config/service_default.yaml`.
**Acceptance:** In a temp dir **outside** the repo: `axiomod init demo --dev --framework-path <repo>` → `cd demo && go mod tidy && go build ./...` succeeds. VERIFY EMPIRICALLY. (The non-`--dev` path becomes verifiable after AX-031 publishes a tag — assert the emitted go.mod contents in a unit test meanwhile.)

### AX-021 · `init`/`generate`: gofmt-clean output + promised files
**Priority:** P0 · **Size:** S · **Depends:** AX-020
**Problem:** Generated `main.go` fails `gofmt -l` (stray leading spaces); `docs/cli-reference.md` says init creates a Dockerfile — it doesn't.
**Steps:**
1. Pipe every generated `.go` file through `go/format.Source` before writing (in `init.go` and `cmd/axiomod/cmd/generate/{module,handler,service}.go`); fix the template whitespace at the source too.
2. Add a Dockerfile to the init scaffold (reuse the dockerize template, base `golang:1.24-alpine`, parameterized binary name) — or, if rejected, fix cli-reference instead; default: generate it.
3. Unit test: render each template, assert `format.Source` round-trips unchanged.
**Acceptance:** `gofmt -l` on a freshly scaffolded project prints nothing; scaffold contains a Dockerfile.

### AX-022 · CI: scaffold smoke test
**Priority:** P0 · **Size:** S · **Depends:** AX-020, AX-021
**Steps:** New CI job: build CLI → `axiomod init smoke --dev --framework-path "$GITHUB_WORKSPACE"` in a temp dir → `go mod tidy && go build ./... && go vet ./...` inside it. This permanently guards the quick start.
**Acceptance:** Job green in CI; intentionally breaking a template locally makes it fail.

### AX-023 · Register the example module in the server
**Priority:** P0 · **Size:** S · **Depends:** AX-010
**Problem:** `example.Module` is commented out in `cmd/axiomod-server/fx_options.go:30-31`; the shipped server exposes no domain API (audit B6).
**Steps:** Import `examples/example` and add `example.Module` to `getModuleOptions()`. Confirm `registerHTTPRoutes`/`registerGRPCServices` invocations in `examples/example/module.go` work against the provided `*fiber.App` and gRPC server. Extend `cmd/axiomod-server/integration_test.go`: boot app, `POST /api/v1/examples/` then `GET /api/v1/examples/:id` round-trip (note the example's auth middleware — see AX-025; if it blocks, pass a Bearer token per its current contract).
**Acceptance:** Boot log shows the example routes; integration test green. VERIFY EMPIRICALLY with curl.

### AX-024 · Example module: complete CRUD over HTTP + gRPC
**Priority:** P0 · **Size:** L · **Depends:** AX-023
**Problem:** Only Create/Get use cases exist; repository already implements Update/Delete/List (audit B6).
**Steps:**
1. New use cases (one per file, `Execute(ctx, input) (output, error)` per `.claude/rules/09-domain-patterns.md`): `update_example.go`, `delete_example.go`, `list_example.go` (List maps `repository.ExampleFilter`, supports limit/offset).
2. HTTP: `PUT /api/v1/examples/:id`, `DELETE /api/v1/examples/:id`, `GET /api/v1/examples/` (query params name/valueType/tag/limit/offset) in `example_handler.go`; map domain errors → status codes via `framework/errors.ToHTTPCode` where applicable.
3. Proto: extend `examples/example/delivery/grpc/example.proto` with `UpdateExample`, `DeleteExample`, `ListExamples` RPCs; regenerate:
   ```bash
   go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
   go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
   protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative examples/example/delivery/grpc/example.proto
   ```
   (If `protoc` is unavailable in the environment, install via package manager or use `buf`; do not hand-edit `*.pb.go`.)
4. Implement the new gRPC methods in `example_grpc_service.go` with `status.Error(framework/errors.ToGRPCCode(err), …)`.
5. Wire new use cases in `module.go`; extend `example_test.go` (table-driven) for all new paths.
**Acceptance:** All five operations succeed over HTTP (curl) and gRPC (`grpcurl`, reflection is on). VERIFY EMPIRICALLY.

### AX-025 · Example auth middleware: real JWT validation
**Priority:** P0 · **Size:** S · **Depends:** AX-023
**Problem:** `examples/example/delivery/http/middleware/middleware.go` "AuthMiddleware" only checks Bearer *format* — a misleading security example in the reference app (audit B6).
**Steps:** Replace with `framework/middleware.AuthMiddleware` (real JWT validation, claims in `c.Locals`) injected via the module, or validate via `auth.JWTService` inside the example middleware. Update `module.go` wiring and tests (generate a token with the test secret).
**Acceptance:** Request without valid JWT → 401; with valid token → 2xx; test proves both.

### AX-026 · Ent as the framework's default ORM (config-selectable raw SQL)
**Priority:** P0 · **Size:** L · **Depends:** AX-010, AX-023
**Directive change (2026-06-12):** originally this task stripped the fake Ent repository; per maintainer direction it now does the opposite — implement Ent **for real** as the framework default, keeping plain `database/sql` available behind a config switch. Supersedes AX-070.
**Problem:** `example_ent_repository.go` is raw `database/sql` with comments admitting it isn't Ent; `platform/ent` referenced in rules/CLAUDE.md doesn't exist (audit B6).
**Steps:**
1. Add `entgo.io/ent` to `go.mod`; create `platform/ent/schema/example.go` (fields mirroring `entity.Example`: string id, name, description, value type/count, JSON tags, timestamps) and run ent codegen into `platform/ent/` (`go generate ./platform/ent/...`).
2. Provide a client constructor (`platform/ent/client.go`): wrap an existing `*sql.DB` from `framework/database` via `entsql.OpenDB(dialect, db)` so pooling/metrics/slow-query plumbing is reused.
3. New config key `database.orm: "ent" | "sql"` (**default `ent`**) — add `ORM` field to `DatabaseConfig` in `framework/config/types.go`, default it in `setDefaults`, document in `configs/service_default.yaml`.
4. Rewrite `examples/example/infrastructure/persistence/example_ent_repository.go` as a real Ent implementation of `repository.ExampleRepository`; keep the raw-SQL implementation as `example_sql_repository.go` (honest comments).
5. Repository selection in `examples/example/module.go`: memory when no DB plugin is enabled (server must boot with zero infra); when postgres/mysql is enabled choose Ent vs SQL by `database.orm`.
6. Restore/keep `platform/ent/schema` + `platform/ent/migrate` exceptions in `architecture-rules.json` and the "repositories may import platform/ent" pattern rule; keep `.claude` references accurate.
7. Tests: Ent repository CRUD against an in-memory SQLite driver if feasible without cgo (`modernc.org/sqlite`), otherwise gate behind `RUN_INTEGRATION_TESTS`; unit-test the selection logic.
**Acceptance:** `database.orm` defaults to `ent`; with `orm: sql` the SQL repository is selected (asserted by test); server still boots with no DB configured; build/tests green.

---

## WAVE 4 — Release engineering (P0)

### AX-030 · goreleaser + tag-triggered release workflow
**Priority:** P0 · **Size:** M · **Depends:** AX-007
**Steps:**
1. Add `.goreleaser.yaml`: two builds (`axiomod` from `./cmd/axiomod`, `axiomod-server` from `./cmd/axiomod-server`), `CGO_ENABLED=0`, targets linux/darwin × amd64/arm64, ldflags injecting `framework/version.{Version,GitCommit,BuildDate}` (copy exact `-X` paths from the `Makefile`), archives + checksums.
2. Add `.github/workflows/release.yml` triggered on `v*` tags: checkout (fetch-depth 0), setup-go, `goreleaser release --clean`.
3. `goreleaser check` and `goreleaser release --snapshot --clean` locally to validate config (do **not** tag).
4. Update `docs/release-checklist.md` to the new flow; keep the `.claude/hooks/enforce-version-sync.sh` guard intact.
**Acceptance:** Snapshot release builds all artifacts locally; workflow lints clean (`actionlint` if available).

### AX-031 · Cut the first real release (maintainer-gated)
**Priority:** P0 · **Size:** S · **Depends:** all other P0 tasks merged
**Steps (agent prepares, maintainer executes the tag/push):** bump via `./scripts/bump-version.sh v0.3.0`; write `docs/release-notes/v0.3.0.md` summarizing the launch-readiness fixes; verify `go.mod` module path tag-compatibility; after tag exists, verify `GOPROXY=direct go get github.com/axiomod/axiomod@v0.3.0` from a scratch module and that non-`--dev` `axiomod init` output now builds (closes the AX-020 loop). Fix README badges that point at nonexistent releases if still broken.
**Acceptance:** Tagged GitHub release with binaries; `go get` works; README badges render.

---

## WAVE 5 — CLI integrity (P1)

### AX-040 · Wire the orphaned `config validate` / `config diff` subcommands
**Priority:** P1 · **Size:** S
**Problem (verified):** `NewConfigValidateCmd()`/diff live in package `cmd/core/config` but nothing attaches them to the parent `configCmd` (`cmd/core/config.go`) — `axiomod config validate` prints help and exits 0 without validating anything; `config` shows as a bare "help topic" in root help.
**Steps:** In `core.NewConfigCmd()`, `configCmd.AddCommand(config.NewConfigValidateCmd(), config.NewConfigDiffCmd())` (import the subpackage). Point `validate` at `configs/` (post-AX-001). Add a CLI test (table-driven, exec the cobra command) asserting `config validate` actually loads YAML and that an invalid file yields exit ≠ 0.
**Acceptance:** `axiomod config --help` lists both subcommands; `config validate` on a corrupted YAML fails. VERIFY EMPIRICALLY.

### AX-041 · `status`: replace the always-healthy stub
**Priority:** P1 · **Size:** S
**Problem:** `cmd/axiomod/cmd/core/status.go` `checkHealth()` unconditionally returns nil (line ~51).
**Steps:** Query `http://<host>:<port>/ready` and `/health` (default `localhost:8080`, `--url` flag to override), print component statuses from the JSON `Response{Status, Components}` shape in `framework/health`, exit 1 when not UP. Reuse/merge with `healthcheck.go` logic — consider making `status` the rich version and `healthcheck` a thin alias; do not leave two diverging implementations.
**Acceptance:** Against a running server: exit 0 + component table; against nothing: exit 1 with clear error. VERIFY EMPIRICALLY.

### AX-042 · `logs`: implement minimally or remove
**Priority:** P1 · **Size:** S
**Problem:** Stub with commented-out kubectl example; documented `--follow` does nothing.
**Steps (recommended):** Remove the command from `root.go` and delete `logs.go`, removing it from `docs/cli-reference.md` (a server framework CLI shouldn't fake log tailing; `docker logs`/`kubectl logs` own this). If maintainer prefers keeping it: implement `docker logs`-shelling with `--container` flag and honest errors. Either way, no silent no-op.
**Acceptance:** Command either gone from `--help` + docs, or demonstrably tails a real container.

### AX-043 · `validator check-api-spec`: real OpenAPI validation
**Priority:** P1 · **Size:** M
**Problem:** `cmd/axiomod/cmd/validator/validator_functions.go:32-45` prints "placeholder in this version" and always succeeds.
**Steps:** Use `github.com/getkin/kin-openapi/openapi3`: `openapi3.NewLoader().LoadFromFile(specPath)` + `doc.Validate(ctx)`; report errors with paths; exit ≠ 0 on invalid. Support JSON and YAML. Table-driven tests with a valid and a broken spec in `testdata/`. Update `docs/validator-guide.md` (it currently claims "spectral" — make it state the real mechanism).
**Acceptance:** Valid spec → 0; spec with a broken `$ref` → non-zero with the error printed.

### AX-044 · `deploy`: stop simulating silently
**Priority:** P1 · **Size:** S
**Problem:** Docker build is real; registry push/k8s apply are print-only theatre (`deploy.go:58-62`).
**Steps:** Print an explicit `[SIMULATED]` prefix on the fake steps, add `--dry-run` (default **true**) so real execution is opt-in, and only attempt real `docker push`/`kubectl apply` when `--dry-run=false` AND the tools/flags (`--registry`, `--kubeconfig`) are provided; otherwise exit with "not configured" guidance. Update cli-reference (`--env` flag documented there doesn't exist — it's a positional arg; align doc and code by adding the flag or fixing the doc; recommended: keep positional, fix doc in AX-050).
**Acceptance:** Default run clearly labels simulation; `--dry-run=false` without config fails loudly.

### AX-045 · `dockerize`: `--tag` flag + correct Go version
**Priority:** P1 · **Size:** S
**Problem:** Docs document `--tag=v1.0.0`; code hardcodes `go-axiomod:latest` and the generated Dockerfile uses `golang:1.21` (repo needs Go 1.24).
**Steps:** Add `--tag` (default `<module-name>:latest` derived from go.mod, not hardcoded `go-axiomod`); template base image `golang:1.24-alpine`, runtime `alpine:3.21`, non-root user — mirror the repo's own Dockerfile structure.
**Acceptance:** `axiomod dockerize --tag=myapp:v1 ` writes a Dockerfile whose `FROM` matches and invokes `docker build -t myapp:v1` (assert the exec args in a test; actual docker run optional).

### AX-046 · `plugin list/install/remove`: fix paths and honesty
**Priority:** P1 · **Size:** S
**Problem:** `plugin list` scans `internal/plugins/` — wrong for this layout (`plugins/`); install/remove only clone/rm-rf and print "manually register".
**Steps:** Scan `plugins/` (walk one level, a dir counts if it contains a `plugin.go` implementing `Name()` — string match is fine); skip `example_plugin` as today. For install/remove, print precise registration instructions naming `cmd/axiomod-server/register_plugins.go` (the real location) instead of generic text. Defer AST auto-registration to AX-074.
**Acceptance:** `axiomod plugin list` in repo root lists ldap, saml, multitenancy, audit, elk, redis, kafka. VERIFY EMPIRICALLY.

### AX-047 · Reconcile remaining flag drift (`generate handler`, `test`, `build`)
**Priority:** P1 · **Size:** S
**Steps:** Decide per flag — implement or de-document (do the doc side in AX-050 in the same PR series):
- `generate handler --type=http|grpc`: implement `--type` with `grpc` emitting a service skeleton + proto stub, or drop from docs. Recommended: drop from docs now, roadmap the grpc generator.
- `test --unit/--integration`: implement as `go test ./tests/unit/...` / `RUN_INTEGRATION_TESTS=true go test ./tests/integration/...` mappings — small and genuinely useful. Recommended: implement.
- `build`: stop hardcoding `./cmd/axiomod-server`; detect `./cmd/<module-name>` from go.mod or accept a positional package path, so it works in scaffolded projects.
**Acceptance:** `--help` output and cli-reference agree for all three commands; `axiomod test --unit` runs only unit tests.

### AX-048 · Bump cobra to clear the "help topcis" typo
**Priority:** P1 · **Size:** XS
**Problem:** Root `--help` prints "Additional help topcis" — the typo is not in repo source; it ships with the pinned cobra version.
**Steps:** `go get github.com/spf13/cobra@latest && go mod tidy`; re-run CLI help to confirm; full DoD gate (cobra bumps occasionally change help layout — eyeball snapshot tests if any).
**Acceptance:** `axiomod --help | grep -i topcis` empty. VERIFY EMPIRICALLY.

---

## WAVE 6 — Docs reconciliation (P1)

> Rule for all Wave-6 tasks: the **binary is the source of truth** (after Wave 5 lands). Regenerate claims from `--help` output and actual config structs — never from memory. Where behavior is still missing, write "planned" honestly and link the roadmap.

### AX-050 · Rewrite `docs/cli-reference.md` from `--help` truth
**Priority:** P1 · **Size:** M · **Depends:** Wave 5
**Steps:** Walk every command with `--help` and document the real surface. Known corrections from the audit: `config view/check` → `config validate/diff`; `policy add p <sub> <obj> <act>` positional syntax (no `--ptype/--v0/--v1/--v2`); no `generate handler --type` (unless AX-047 added it); `test` flags per AX-047; `dockerize --tag` per AX-045; `deploy` positional env; `logs` removed (AX-042); document `interactive`, `validator domain/static-analysis/static-check/check-docs/standards-check` which exist but were undocumented or under-documented; remove `validator run` ghosts. Add the validator exit-code contract.
**Acceptance:** Spot-check script: every documented command/flag string appears in the corresponding `--help` output (write a small shell loop in the PR description proving it).

### AX-051 · Fix `docs/deployment-guide.md` (env vars, config keys, k8s)
**Priority:** P1 · **Size:** M · **Depends:** AX-001/002/003/004, AX-066
**Steps:** Replace the entire env-var section with the real scheme (`APP_` prefix + Viper key mapping: `APP_HTTP_PORT`, `APP_DATABASE_HOST`, `APP_DATABASE_PASSWORD`, `APP_AUTH_JWT_SECRETKEY`, …) — cross-check each against `framework/config/types.go` field names. Fix the config example: remove top-level `kafka:` (it lives under `plugins.settings.kafka`), `auth.provider/jwtSecret/jwtDuration` → `auth.jwt.secretKey/tokenDuration`, `tracingExporterURL` → `tracingUrl`, `plugins.enabled` list → map of bools, `plugins.config` → `plugins.settings`. Point the Kubernetes section at the real manifests from AX-066 (or cut it if AX-066 is descoped).
**Acceptance:** Every YAML key in the guide exists in `configs/service_default.yaml` or `types.go`; every env var round-trips (`APP_HTTP_PORT=18080 ./bin/axiomod-server` binds 18080 — VERIFY EMPIRICALLY).

### AX-052 · Fix the plugin guide + `example_plugin`
**Priority:** P1 · **Size:** S
**Problem:** `docs/plugin-development-guide.md` teaches a 2-arg `Initialize(config, *zap.Logger)`; the real interface (`plugins/plugin.go`) takes 5 args. `plugins/example_plugin/example_plugin.go` implements the outdated shape, doubling the misinformation.
**Steps:** Rewrite `example_plugin` to implement the real `Plugin` interface (5-arg Initialize, health-check registration, lifecycle per `.claude/rules/08-plugin-system.md`) with heavy comments — it is the teaching artifact; add a test. Rewrite the guide from it: registration via `RegisterNewPlugins` in `cmd/axiomod-server/register_plugins.go` (extended) or `registerBuiltInPlugins` (built-ins), `plugins.enabled` as a **map**, settings under `plugins.settings.<name>`, env override via `APP_PLUGINS_...`. Delete the fictional `PLUGINS_ENABLED` env-var section.
**Acceptance:** Guide's example compiles if pasted (prove by making `example_plugin` literally be the guide's code); interface signature in the doc matches `plugins/plugin.go` verbatim.

### AX-053 · Fix `docs/api-reference.md`
**Priority:** P1 · **Size:** S
**Steps:** Remove/replace the claim that proto contracts live in `api/` — either create `api/proto/` and move `examples/example/delivery/grpc/example.proto` there (updating `go_package` + regen + generator docs), or document the actual convention (`<domain>/delivery/grpc/*.proto`). Recommended: document the actual convention now; an `api/` dir migration can ride with AX-071. State plainly that OpenAPI specs are not yet generated (link roadmap), and document the example module's real endpoints (post-AX-024) as the worked example.
**Acceptance:** Every path referenced in the doc exists in the repo.

### AX-054 · Fix observability / database / auth-security guides
**Priority:** P1 · **Size:** M · **Depends:** AX-004, AX-005
**Steps:**
- `observability-guide.md`: env vars → `APP_OBSERVABILITY_*` forms; `tracingSamplingRatio` → `tracingSamplerRatio`; `tracingExporterURL` → `tracingUrl`; metrics location per AX-004 outcome.
- `database-guide.md`: complete the truncated §4 ("Plugins") — document the postgres/mysql plugins' enable/settings flow; align field names with `configs/service_default.yaml` (`user` vs `username`); fix the self-contradictory `connMaxLifetime: 300 # minutes (default 5)` line; state migrate driver support truthfully (post-AX-005).
- `auth-security-guide.md`: verify the exact settings keys the keycloak plugin reads (`grep -n "settings\[" plugins/builtin_plugins.go` region) and make the YAML examples match (camelCase per `.claude/rules/11-config-system.md` if that's what the code reads — fix code or doc to the rules' convention, prefer fixing the doc unless code violates its own convention).
**Acceptance:** Every config key and env var in all three guides resolves against code; one empirical env-var override demonstrated per guide.

### AX-055 · `docs/testing-guide.md` vs CI honesty
**Priority:** P1 · **Size:** XS · **Depends:** AX-062 (or do jointly)
**Steps:** Replace ">80% coverage gate" with the actual enforced number + the published ramp plan (28 → 50 → 80, per AX-062); remove the Go 1.24/1.25 matrix claim or implement a matrix in `ci.yml` (recommended: implement `strategy.matrix.go: ['1.24.x', '1.25.x']` — cheap and true).
**Acceptance:** Guide statements match `ci.yml` exactly.

### AX-056 · Rewrite `docs/readiness-assessment.md` honestly
**Priority:** P1 · **Size:** S · **Depends:** end of Wave 3 (so the rewrite reflects fixes)
**Steps:** Re-run the audit checklist (build/test/boot/validator/quick-start) and rewrite statuses with evidence per row; downgrade anything not verified; date it; link `docs/audit/2026-06-12-launch-readiness/feature-implementation-audit.md` as methodology. Keep 🟢 only for items with a passing verification command listed inline.
**Acceptance:** Every 🟢 row cites a command an evaluator can run.

### AX-057 · Root `README.md` polish
**Priority:** P1 · **Size:** XS
**Steps:** Fix the artifact "enterprise Go development. functionality."; move Multi-Tenancy out of the "coming next" list (it ships — link the plugin row instead); ensure the Quick Start matches AX-020 reality (including `--dev` note for in-repo use); badges audit (Release badge only after AX-031).
**Acceptance:** README quick start executes verbatim, copy-paste, on a clean machine (post-AX-031) or with the documented `--dev` flag (pre-tag).

---

## WAVE 7 — Quality bar (P1)

### AX-060 · Tests for `framework/errors` (currently zero)
**Priority:** P1 · **Size:** M
**Steps:** Table-driven coverage of: `New/Wrap/WithCode/WithMetadata` (including metadata cloning/immutability), all shorthand constructors, `GetCode/GetMetadata/GetStack`, `Is/As` unwrap chains, full `ToHTTPCode` and `ToGRPCCode` mapping tables (every code constant → expected status), stack capture skipping runtime frames, nil-error edge cases. Add the benchmark stub per `.claude/rules/04-testing.md`. Target >85% for this package.
**Acceptance:** `go test -cover ./framework/errors/` ≥ 85%.

### AX-061 · Thin-package test backfill
**Priority:** P1 · **Size:** L (split into one PR per package)
**Targets & focus:**
- `framework/database`: DSN building (both drivers), redaction, slow-query logging trigger (fake clock or low threshold with sqlmock), `WithTransaction` commit path (only rollback is tested today).
- `framework/health`: RegisterCheck/RunChecks aggregation, DOWN propagation, `Handler()` HTTP codes, background-check loop start/stop.
- `framework/circuitbreaker`: full state walk Closed→Open→HalfOpen→Closed, HalfOpenLimit, reset-timeout reopen, concurrent `Execute` under `-race`.
- `framework/validation`: custom validator registration, message mapping per tag, JSON tag names.
- `framework/worker`: interval execution (short intervals), per-job timeout firing, StopJob/StopAll idempotency, Shutdown drains.
- `framework/kafka`: config defaulting, handler registry dispatch with a fake consumer-group session (no broker).
- `framework/utils`, `framework/version`, `framework/router` (router only if AX-063 keeps it).
**Acceptance:** Each package ≥ 60% in its PR; no test uses live external services un-gated.

### AX-062 · Raise the CI coverage gate
**Priority:** P1 · **Size:** XS · **Depends:** AX-060, AX-061
**Steps:** Bump the threshold in `.github/workflows/ci.yml` from 28 to 50 once Wave-7 tests land; leave a tracked TODO (issue) for 80. Sync `docs/testing-guide.md` (AX-055).
**Acceptance:** CI green at the new threshold on main.

### AX-063 · Delete dead weight: `framework/router`, `scripts/run_tests.go`
**Priority:** P1 · **Size:** S
**Steps:** Confirm `framework/router` consumers: `grep -rn "framework/router"` (known: `cmd/axiomod-server/framework_test.go` provides it). It duplicates `platform/server`'s Fiber setup and is untested — remove the package and fix the test, OR (if maintainer objects) wire it as the actual server router and test it. Recommended: remove. Also delete `scripts/run_tests.go` (redundant with `make test`); check `scripts/Makefile` references.
**Acceptance:** Build/tests green with package removed; no doc references remain.

### AX-064 · OSS hygiene files
**Priority:** P1 · **Size:** S
**Steps:** Add root `CONTRIBUTING.md` (build/test/lint/validator workflow, PR checklist mirroring `.claude/CLAUDE.md` post-implementation checklist), `SECURITY.md` (private reporting contact, supported-versions table), `CODE_OF_CONDUCT.md` (Contributor Covenant 2.1), `CHANGELOG.md` (Keep-a-Changelog format, seeded from `docs/release-notes/`). Link all four from README; update `docs/release-checklist.md` to require a CHANGELOG entry.
**Acceptance:** Files exist, linked, lint-clean markdown.

### AX-065 · Dockerfile HEALTHCHECK
Folded into **AX-003** — listed here only so the wave checklist is complete. No separate PR.

### AX-066 · Ship the Kubernetes assets the docs describe
**Priority:** P1 · **Size:** M · **Depends:** AX-003/AX-004 (final ports)
**Steps:** Create `deploy/kubernetes/{deployment.yaml,service.yaml,configmap.yaml}` matching `docs/deployment-guide.md`'s described probes (`/ready` readiness 5s/10s, `/live` liveness 15s/20s), resources, 3 replicas, ports; ConfigMap mounts `service_default.yaml`; Secret refs for `APP_DATABASE_PASSWORD`/`APP_AUTH_JWT_SECRETKEY`. Validate with `kubectl apply --dry-run=client -f` (or `kubeconform` if no cluster). Update the guide's paths (`kubernetes/…` → `deploy/kubernetes/…`). Helm chart stays P2 (AX-073).
**Acceptance:** Manifests pass dry-run validation; doc paths resolve.

---

## WAVE 8 — Post-launch competitive parity (P2 — design briefs, not yet scheduled)

| ID | Task | Notes |
|---|---|---|
| **AX-070** | ~~Real Ent integration~~ **Superseded by AX-026** (pulled forward to Wave 3 per maintainer directive). | — |
| **AX-071** | OpenAPI story: swaggo annotations on example handlers, `make openapi` target, publish `api/openapi.yaml`, validate in CI via AX-043's validator. Pairs with roadmap "generate from-spec". | M |
| **AX-072** | Generators emit tests: `generate module/handler/service` scaffold matching `_test.go` files (table-driven templates). | M |
| **AX-073** | Helm chart (`deploy/helm/axiomod/`) with values for ports/probes/secrets; docs site (mkdocs-material or Docusaurus) publishing `docs/` via GitHub Pages. | L |
| **AX-074** | Roadmap features per `docs/roadmap/*` (already honestly scoped there): CLI gatekeeper (SARIF/JUnit output), binary plugin protocol (`axiomod-*` on PATH), Vault provider, mTLS for gRPC, Redis rate limiting, outbox + DLQ, `migrate --dry-run` / `migrate test`. | Each L; spec before code. |

---

## Progress tracker

Status as of 2026-06-12 (branch `claude/hopeful-cray-yoivk5`):

- [x] **W1:** AX-001 · AX-002 · AX-003 · AX-004 · AX-005 · AX-006 · AX-007 · AX-008 — all done
- [x] **W2:** AX-010 · AX-011 · AX-012 — done; validator: 94 violations → 0, enforced in CI
- [x] **W3:** AX-020 · AX-021 · AX-022 · AX-023 · AX-024 · AX-025 · AX-026 — done; Ent is the default ORM (`database.orm: ent|sql`)
- [x] **W4:** AX-030 done · **AX-031 remains maintainer-gated** (tag v0.3.0 per release checklist)
- [x] **W5:** AX-040 · AX-041 · AX-042 (logs removed) · AX-043 · AX-044 · AX-045 · AX-046 · AX-047 · AX-048 — done
- [x] **W6:** AX-050 · AX-051 · AX-052 · AX-053 · AX-054 · AX-055 · AX-056 · AX-057 — done
- [x] **W7:** AX-060 (errors 94% cov) · AX-061 (backfill complete: database incl. driver-registration + DSN tables + real commit/rollback, health incl. handler/background, circuitbreaker full state walk, validation messages/custom validators, worker timeout/lifecycle, utils, version; total coverage 65.1%) · AX-062 (gate 50%) · AX-063 · AX-064 · AX-066 — done
- [ ] **W8 (P2):** ~~AX-070~~ (superseded by AX-026, done) · AX-071 · AX-072 · AX-073 · AX-074 — not scheduled

**Additional fixes discovered during execution** (not in the original plan):
- SQL drivers were registered only in the migrate CLI — the server binary
  could not open any database; now registered in `framework/database`.
- `server.Module`/`grpc.Module` did not provide `*fiber.App`/`*grpc.Server`,
  so the documented domain-wiring pattern could never resolve; fixed.
- Plugin settings keys are now normalized (Viper lowercases YAML map keys;
  camelCase settings like `elasticsearchUrl`, `bindDn`, `clientId` were
  silently ignored at runtime); keycloak accepts camelCase + legacy keys.
- OIDC made opt-in (idle when `issuerUrl` empty) instead of erroring at boot.

**Launch gate:** W1–W4 merged ✅ + AX-031 tagged (pending, maintainer) +
`docs/readiness-assessment.md` regenerated ✅ (Releases row stays 🔴 until
the tag exists).
