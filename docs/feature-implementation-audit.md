# Axiomod — Documented vs. Implemented Feature Audit

**Date:** 2026-06-12
**Scope:** Every feature claim in `README.md` and `docs/` verified against the codebase on branch `claude/hopeful-cray-yoivk5` (HEAD `eb9ac98`), plus a repo/structure assessment against the bar set by Spring Boot, Laravel, .NET, and Next.js.
**Method:** Full read of all 29 docs files; source audit of all 22 `framework/`+`platform/` packages, all 13 plugins, all 33 CLI commands; empirical verification: `go build ./...`, `go vet`, `gofmt -l`, `go test -race ./...`, CLI binary exercised (`--help`, `init`, `validator architecture`, flag checks), server booted and observed, scaffolded project compiled.

---

## 1. Verdict

The framework core is **substantially real** — this is not vaporware. Auth (JWT/OIDC-with-JWKS/Casbin), dual-protocol serving, observability, resilience, Kafka, the plugin registry, and most of the CLI are genuine, working implementations, and the full test suite passes with `-race`.

However, the project is **not launch-ready as a product**. The single most important user journey — *install → scaffold → run* — is broken end-to-end, the config system silently loads the wrong file (or none, in Docker), the flagship architecture validator fails on the framework's own code with 94 violations, MySQL support is advertised but has no driver dependency, and roughly a third of the CLI reference documents flags and commands that do not exist. The documented architecture (platform may import framework) is the inverse of the actual dependency graph.

---

## 2. Empirical results (this audit, this machine)

| Check | Result |
|---|---|
| `go build ./...` | ✅ PASS |
| `go vet ./...` | ✅ PASS |
| `gofmt -l .` | ✅ clean |
| `go test -race ./...` | ✅ ALL PASS (31 packages with tests) |
| Server boot (`cmd/axiomod-server`) | ✅ boots; HTTP :8080 (14 handlers), graceful shutdown works — ⚠️ but gRPC bound to **:50051**, not the documented/configured **:9090** (see Blocker B2) |
| `axiomod validator architecture` (self) | ❌ **FAILS — 94 violations** (see Blocker B3) |
| `axiomod init` → `go mod tidy` → build | ❌ **FAILS** (see Blocker B1) |

---

## 3. README feature table — verification matrix

| README claim | Status | Evidence / notes |
|---|---|---|
| Clean Architecture | ⚠️ Partial | Example module follows it, but reference impl is incomplete (no Update/Delete/List use cases) and **not registered** in `cmd/axiomod-server/fx_options.go` (commented out). The framework's own layering contradicts its documented rules (§7). |
| fx Dependency Injection | ✅ Implemented | `var Module` pattern across observability, middleware, auth, health, grpc, server, plugins, worker; `framework/di` builder helpers; assembly in `fx_options.go`. |
| Config Management (Viper) | ⚠️ Implemented but defective | Full Provider interface, `APP_` env, `WatchConfig`. **Defect:** loader never searches `configs/` where the documented default file lives (Blocker B2). |
| Fiber v2 HTTP | ✅ Implemented | `platform/server`, middleware order recover→cors→compress→logger→metrics→tracing, `/live` `/ready` `/health` `/metrics`. |
| gRPC | ✅ Implemented | `framework/grpc`: TLS, health service, reflection, keepalive, graceful stop; metrics/tracing/RBAC/timeout/recovery interceptors. |
| Middleware chains | ✅ Implemented | All 8 documented middleware exist with struct-with-`Handle()` pattern (Logging, Auth, Role, Timeout, Recovery, Metrics, Tracing, RBAC). |
| Validators | ✅ Implemented | `framework/validation` wraps validator/v10 with friendly messages. |
| **MySQL plugin** | ❌ **Broken** | `go.mod` contains **no MySQL driver** (only `lib/pq`). `sql.Open("mysql", …)` in `framework/database/database.go:85` will fail with `unknown driver`. The advertised MySQL support cannot work. |
| PostgreSQL plugin | ✅ Implemented | Pooling, health check, slow-query detection, transactions. |
| Connection pooling / transactions | ✅ Implemented | `WithTransaction`, pool limits, DSN redaction in logs. |
| JWT | ✅ Implemented | HS256, ≥32-byte secret enforced (`auth.ValidateSecretKey`). |
| OIDC (Keycloak) | ✅ Implemented | Real discovery + **JWKS signature verification** via `keyfunc/v3` with background refresh (ADR-003 honored). |
| LDAP | ✅ Implemented | Genuine bind-search-bind flow (`plugins/auth/ldap`), health check, TLS option. |
| SAML 2.0 | ✅ Implemented | `crewjam/saml` SP: IdP metadata (URL or file), signing cert support. |
| RBAC (Casbin) | ⚠️ Implemented but unusable OOTB | Real Casbin v2 wrapper + Fiber middleware + gRPC interceptor. **But** the model/policy files referenced by `configs/service_default.yaml` (`rbac_model.conf`, `rbac_policy.csv`) **do not exist in the repo**, so RBAC and the `axiomod policy` commands fail out of the box. |
| Health probes | ✅ Implemented | `/live`, `/ready`, `/health`; `RegisterCheck`; background checks. |
| Prometheus metrics | ✅ Implemented | All 5 documented vectors (HTTP/gRPC counters+histograms, DBQueryDuration). ⚠️ Metrics port drifts across sources: docs+CLAUDE.md say 9100, `configs/service_default.yaml`+Dockerfile say 9091, compose maps 9100. |
| OpenTelemetry traces | ✅ Implemented | jaeger / otlp (TLS or insecure) / stdout exporters, configurable sampler ratio. |
| Kafka producers/consumers | ✅ Implemented | Sarama SyncProducer + ConsumerGroup, fx lifecycle module. |
| Background worker pools | ✅ Implemented | Interval jobs with per-job timeout, graceful shutdown. |
| README Roadmap section | ⚠️ Stale | Lists "Multi-Tenancy Support" as upcoming, but the multitenancy plugin **already ships and works**. Also contains a copy-editing artifact: "…enterprise Go development. functionality." |

## 4. Plugin matrix (13 advertised)

| Plugin | Verdict | Notes |
|---|---|---|
| postgres | ✅ Real | via `database.Connect` |
| **mysql** | ❌ **Cannot work** | no driver dependency in `go.mod` |
| jwt | ✅ Real | secret validation, configurable duration |
| keycloak | ✅ Real | OIDC discovery + verification via `framework/auth` |
| casdoor | ✅ Real | OIDC-based |
| casbin | ⚠️ Real code, missing assets | model/policy files absent from repo |
| ldap | ✅ Real | bind-search-bind |
| saml | ✅ Real | SP, IdP metadata, AuthnRequests |
| multitenancy | ✅ Real | header → `c.Locals("tenant_id")`, optional required-mode |
| auditing | ✅ Real | structured log + JSON-lines file sink |
| elk | ✅ Real | buffered bulk shipping, flush loop |
| redis | ✅ Real | go-redis client + PING health check |
| kafka | ✅ Real | producer on `framework/kafka` |

All 13 have passing test files. Registry semantics (fail-fast on enabled-but-unregistered, Initialize→Start in `StartAll`, Stop on fx `OnStop`) match docs.

## 5. CLI matrix (33 commands)

Verdicts: **REAL** (does what docs say) / **PARTIAL** / **STUB**.

| Command | Verdict | Notes |
|---|---|---|
| `init` | PARTIAL | Scaffolds files, but output **does not build** (Blocker B1); prints next-steps referencing `framework/config/service_default.yaml` while it generates `config/service_default.yaml`; docs claim it generates a Dockerfile — it does not; generated `main.go` is not gofmt-clean. |
| `generate module` | REAL | 8-package Clean Architecture scaffold into `examples/<name>/`. No tests generated. |
| `generate handler` / `generate service` | REAL | 2 files each. Documented `--type=http` flag **does not exist**. |
| `migrate create/up/down/force/version` | REAL | Genuine golang-migrate integration — **PostgreSQL only** (`migrate/utils.go` rejects other drivers; docs imply MySQL works). |
| `policy add/list/remove` | REAL | Real Casbin; **but** docs show `--ptype/--v0/--v1/--v2` flags — actual syntax is positional (`policy add p <sub> <obj> <act>`). Fails OOTB due to missing model/policy files. |
| `validator architecture` | REAL | AST-based import analysis — fails on own repo (B3). |
| `validator domain` | REAL | AST cross-domain check. |
| `validator naming` | REAL | Go/API/SQL naming conventions, `--fix`, `--json`. |
| `validator security` | REAL | runs gosec (auto-install). |
| `validator static-analysis` / `static-check` | REAL | vet + staticcheck + gosec. |
| `validator check-docs` | REAL (heuristic) | git-diff: warns if code changed without docs. |
| `validator check-api-spec` | **STUB** | `validator_functions.go:41` prints "API specification validation is a placeholder in this version" and always succeeds. |
| `config validate` / `config diff` | REAL | But docs document `config view` / `config check`, which **don't exist**; `config` also isn't wired as a full subcommand (appears under "Additional help topcis" — note the typo — in root help). |
| `build`, `test`, `lint`, `fmt` | REAL | Wrappers. Documented `test --unit/--integration` flags **don't exist**. `build` hardcodes the axiomod-server path (wrong for user projects). |
| `dockerize` | REAL | Generates Dockerfile + builds. Documented `--tag` flag **doesn't exist**; image name hardcoded `go-axiomod:latest`; generated Dockerfile uses golang:1.21 (repo requires Go 1.24). |
| `deploy` | PARTIAL | Docker build real; registry push / k8s apply are **print-only simulation**. |
| `status` | **STUB** | `checkHealth()` unconditionally returns nil — always reports healthy. |
| `logs` | **STUB** | Prints placeholder; documented `--follow` does nothing. |
| `healthcheck` | REAL | GET `http://localhost:8080/health`. |
| `plugin install/list/remove` | PARTIAL | install = git clone only; remove = rm -rf only; both print "manually register/unregister" — no code wiring. `list` scans `internal/plugins/`, which is the **wrong directory** for this repo layout (`plugins/`). |
| `interactive` | REAL | working REPL. |
| `version` | REAL | ldflags-injected version info. |
| `validator run` (developer-guide) | ❌ doesn't exist | doc drift. |

## 6. Framework/platform packages — implementation & test depth

All 22 packages are real implementations; none contain TODO/panic stubs. Test depth is the issue:

| Coverage | Packages |
|---|---|
| Solid | resilience (17 tests), grpc (23), client (13), cache (12), crypto (11), events (11), observability (11), middleware (12+), server (11), di (9) |
| Thin (1–4 tests) | auth, config, circuitbreaker, database (1), health (1), kafka (2), validation (1), worker (2) |
| **Zero tests** | **errors** (the error-handling backbone), **router**, **utils**, **version** |

CI gates coverage at **28%** (`ci.yml`), while `docs/testing-guide.md` and `.claude` rules promise **>80%** and a "coverage gate that blocks PRs below 80%". The testing guide also claims a Go 1.24/1.25 version matrix; CI tests a single version from `go.mod`.

---

## 7. Launch blockers (P0)

### B1. The Quick Start does not work
`README.md` (and `cli-reference.md`, `developer-guide.md`) promise: `axiomod init my-service` → `go mod tidy` → run. Reality:

1. Generated `go.mod` contains `replace github.com/axiomod/axiomod => ../` — only meaningful if you scaffold **inside the framework checkout**. Anywhere else, resolution fails immediately.
2. Even with a corrected replace path, `go mod tidy` **fails**: the generated `go.mod` pins no versions, so resolution pulls latest-everything and dies on `cloud.google.com/go/compute/metadata: ambiguous import`.
3. The framework cannot be consumed the normal Go way (`go get github.com/axiomod/axiomod@v0.2.0`) because **no tags exist** (B4).

*An evaluator following the README gets a broken project on day one.* Fix: publish a tagged release, generate a `require` block with pinned versions, drop the replace hack (or gate it behind a `--dev` flag).

### B2. Config system loads the wrong file — or none
- The repo ships its documented config at `configs/service_default.yaml`, but `framework/config/config.go:76-90` searches `.`, `config/`, `../config/`, `../../config/`, `framework/config/`, `/etc/app` — **never `configs/`**.
- A second, **divergent** `framework/config/service_default.yaml` exists (different app name/env, missing the entire observability and database sections, `grpc.port: 50051` vs the documented 9090) and is what actually loads. Empirically confirmed: the booted server announced gRPC on `0.0.0.0:50051` while `configs/service_default.yaml`, the Dockerfile (`EXPOSE 9090`), and docker-compose (`9090:9090`) all say 9090.
- In the Docker image, `configs/` is copied to `/app/configs/` and `ENTRYPOINT` passes no `-config` flag — so the container finds **no config file at all** and built-in `SetDefault` calls are commented out (`config.go:300-302`).

*Consequences: the shipped container's advertised ports are wrong, and operators' edits to the documented config file are silently ignored.* Fix: one canonical default config, add `configs/` to search paths (or `-config configs/service_default.yaml` in ENTRYPOINT), delete the stale copy, restore code-level defaults.

### B3. The flagship governance feature fails on the framework itself
`axiomod validator architecture` → **94 violations** across 119 files (e.g., `plugins/plugin.go:10: plugins imports platform/observability (not allowed)` — yet `architecture-rules.json` and all docs explicitly allow `plugins/* → platform/*`). Top violators: `cmd/axiomod-server` (18), `plugins` (8), `examples/example` (8), `framework/grpc` (7).

Two root causes:
1. **The real dependency direction contradicts the documented one.** Docs/CLAUDE.md state `platform/*` may import `framework/*`. In reality, 7+ `framework/` packages (grpc, database, auth, middleware, kafka, router, worker) import `platform/observability`. Either observability belongs in `framework/` (or its own leaf layer), or the documented layering must be inverted.
2. The validator/rules don't encode the documented allowances (`plugins→platform`, `cmd→everything`).

*Selling "architecture enforcement" while failing your own validator is a credibility problem for an enterprise launch.* Fix: decide the true layering, move/document accordingly, make `validator architecture` green, then wire it into CI as the docs claim.

### B4. Release engineering doesn't exist yet
- `git tag -l` is **empty** (fresh clone — remote has no tags), despite `docs/release-notes/v0.1.0.md` and `v0.2.0.md`, a release checklist, a version-sync hook, and README badges for GitHub Releases.
- `v0.1.0.md` is internally titled "**v1.0.0** - Enterprise Ready".
- Two compiled binaries (`axiomod`, `axiomod-server`, ~62 MB combined, **macOS Mach-O x86_64**) are committed at the repo root of this Linux-server framework.

Fix: remove binaries from git history (BFG/filter-repo — they bloat every clone), tag v0.2.0, create GitHub Releases (goreleaser is the ecosystem standard), align release-note titles.

### B5. Advertised MySQL support cannot work
No MySQL driver in `go.mod`; only `lib/pq`. Enabling the `mysql` plugin fails at `sql.Open`. The migrate CLI is additionally hardcoded PostgreSQL-only (`cmd/axiomod/cmd/migrate/utils.go`). Either add `go-sql-driver/mysql` (+ migrate support + CI coverage) or remove MySQL from README/docs until real.

### B6. The reference implementation is incomplete and disconnected
- `examples/example` implements only **Create + Get** use cases; Update/Delete/List exist in the repository layer but have no use cases, HTTP routes, or gRPC methods.
- `example.Module` is **commented out** in `fx_options.go:30-31` — the shipped server exposes no domain API; "14 handlers" are infra endpoints only.
- `example_ent_repository.go` is **not Ent** — raw `database/sql` with a comment admitting "In a real implementation, we would use the Ent ORM". `platform/ent/{schema,migrate}` are referenced by `architecture-rules.json` and CLAUDE.md but **don't exist**.
- `delivery/http/middleware.AuthMiddleware` in the example only checks Bearer-token *format* — a poor security example for a reference app.

---

## 8. Documentation drift inventory (P1)

Doc-by-doc list of claims that don't match the code (beyond the blockers above):

**cli-reference.md** — `config view`/`config check` (don't exist; actual: `validate`/`diff`); `policy add --ptype/--v0/--v1/--v2` (actual: positional `p|g` args); `generate handler --type=http`; `test --unit/--integration`; `dockerize --tag`; `logs --follow` (stub); `init` "creates Dockerfile" (it doesn't) and "pkg/" (not generated).

**developer-guide.md** — `axiomod validator run` (no such command); `make generate` (no such Makefile target; targets are: all, build, build-cli, clean, test, deps, lint, fmt, docker, help).

**deployment-guide.md** — Environment variables are wrong throughout: documents `APP_NAME`, `HTTP_PORT`, `DB_HOST`, `DB_PASSWORD`… The real scheme is the `APP_` prefix with Viper key mapping (`APP_HTTP_PORT`, `APP_DATABASE_HOST`, `APP_DATABASE_PASSWORD`). Config example uses keys that don't exist in `framework/config/types.go`: top-level `kafka:` section (Kafka is configured under `plugins.settings.kafka`), `auth.provider/jwtSecret/jwtDuration` (actual: `auth.jwt.secretKey/tokenDuration`), `observability.tracingServiceName/tracingExporterURL` (actual: `tracingUrl`), `plugins.enabled` as a YAML *list* (actual: a *map* of bools), `plugins.config` (actual: `plugins.settings`). References `kubernetes/deployment.yaml` and `kubernetes/service.yaml` — **no Kubernetes manifests exist anywhere in the repo**.

**plugin-development-guide.md** — Documents a 2-argument `Initialize(config, *zap.Logger)` interface; the real interface (`plugins/plugin.go`) takes 5 args (config, logger, metrics, cfg, health). Registration instructions point to `registerBuiltInPlugins` in `plugins/plugin.go`; extended plugins actually register via `RegisterNewPlugins` in `cmd/axiomod-server/register_plugins.go`. Shows `plugins.enabled` as a list and a `PLUGINS_ENABLED` env var — neither matches the implementation.

**api-reference.md** — Says proto contracts live in an `api/` directory; no `api/` directory exists (the only `.proto` is `examples/example/delivery/grpc/example.proto`). No OpenAPI/Swagger spec exists anywhere despite the recommendation.

**observability-guide.md** — Documents `LOG_LEVEL`, `METRICS_PORT`, `TRACING_ENABLED` env vars (real: `APP_OBSERVABILITY_*`); `tracingSamplingRatio` (real key: `tracingSamplerRatio`); `tracingExporterURL` (real: `tracingUrl`); metrics port 9100 vs configured 9091.

**database-guide.md** — Section 4 ("Plugins") is truncated mid-sentence; documents MySQL flows that can't work (B5); `connMaxLifetime: 300 # minutes (default 5)` is self-contradictory.

**testing-guide.md** — ">80% coverage gate", "Go 1.24/1.25 matrix" — CI implements neither (28% gate, single Go version).

**auth-security-guide.md** — Keycloak settings use snake_case keys (`client_id`, `client_secret`) while the config system convention (and rules/11) is camelCase; needs alignment with the actual keys read by the plugin.

**readiness-assessment.md** — Declares "Stable / Production Ready" and 🟢 across the board. Given B1–B6, this document overstates reality and should be rewritten before launch; an enterprise evaluator who diffs it against behavior will lose trust.

**docs/README.md + root README** — Link set is fine; root README's Roadmap teases multi-tenancy which already ships; badge row references releases/contributors that don't exist yet.

**Roadmap docs** (`roadmap/*.md`) — Accurate and honest (all items correctly marked as planned: OpenAPI scaffolding, gatekeeper, workspaces, Vault, mTLS, outbox/DLQ, rate limiting, SLOs, pprof endpoints…). These are the *right* gaps to name; no false claims found here.

---

## 9. Project & repo structure assessment

**What matches enterprise-framework expectations (keep and showcase):**
- Clear 4-ring layout (`framework/` → `platform/` → `plugins/` → `cmd/`), import-rule file, AST-based self-validation tooling — this is a genuine differentiator; none of Spring/Laravel ship an architecture linter.
- Consistent fx module convention, plugin lifecycle with fail-fast semantics, struct-with-Handle middleware, table-driven tests, ADRs, env-overlay configs, multi-stage non-root Dockerfile, CI with fmt/vet/lint/race/CodeQL, MIT license, a serious `.claude/` engineering-agent setup.

**Where the structure falls short of the Spring Boot / Laravel / .NET / Next.js bar:**

| Dimension | Those frameworks | Axiomod today |
|---|---|---|
| Install story | `spring init` / `composer create-project` / `dotnet new` / `create-next-app` produce a running app in minutes | `axiomod init` output doesn't compile (B1); framework not go-get-able (no tags) |
| Versioned releases | Semantic releases, LTS channels | Zero tags, committed binaries, release notes mislabeled |
| Config | One canonical, layered config story | Two divergent default files; search path misses the documented one (B2) |
| Persistence | First-class ORM (JPA/Eloquent/EF) | Ent referenced everywhere, implemented nowhere; raw-SQL "ent" stub; MySQL claim without driver |
| API contracts | OpenAPI generation built-in (springdoc, NestJS-style) | No `api/` dir, no OpenAPI, one example proto; `check-api-spec` is a placeholder |
| Reference app | Petclinic / Laravel Bootcamp / eShop | Example module: 2 of 5 CRUD ops, not wired into the server |
| Deploy assets | Helm charts, buildpacks, official images | Docs reference k8s manifests that don't exist; container loads no config |
| Test culture | Framework test kits, high published coverage | 28% CI gate vs 80% promise; zero tests on `framework/errors` |
| OSS hygiene | CONTRIBUTING, SECURITY, CoC, CHANGELOG, docs site | All four files missing; docs only in-repo markdown |
| Structural consistency | Layering enforced in CI | Own validator fails (94), framework→platform inversion |

**Specific structural oddities to fix:**
1. Committed macOS binaries at repo root (remove from history).
2. `framework/router` duplicates `platform/server`'s job and is unused — fold or delete.
3. `examples/dummy-api` is a nested Go module with its own replace — fine, but document it; `tests/unit` + `tests/integration` partially duplicate co-located package tests (config/logger tests are trivial).
4. `plugins/example_plugin` implements a *different, outdated* interface (1-arg Initialize) than `plugins/plugin.go` — actively misleads plugin authors (and the plugin guide copies it).
5. `platform/ent/schema|migrate` exceptions in `architecture-rules.json` reference a nonexistent tree.
6. `scripts/run_tests.go` is redundant with `make test`.

---

## 10. Prioritized punch list

**P0 — must fix before any launch announcement**
1. Make the quick start work: pinned `require` in generated go.mod, no `replace` by default, tagged release so `go get` works; CI job that scaffolds + builds a project on every push.
2. Single canonical default config; add `configs/` to search paths or pass `-config` in the Docker ENTRYPOINT; delete `framework/config/service_default.yaml` or regenerate it from one source; restore code defaults; reconcile the gRPC (9090/50051) and metrics (9091/9100) ports everywhere.
3. Resolve the layering contradiction (move observability or flip the documented rule), update `architecture-rules.json`, get `validator architecture` to 0 violations, add it to CI.
4. Remove committed binaries from git history; tag and publish v0.2.0 (goreleaser); fix `v0.1.0.md` title.
5. Either ship MySQL for real (driver + migrate + tests) or remove every MySQL claim.
6. Complete and register the example module (full CRUD over HTTP+gRPC, real auth middleware) — it's the shop window.
7. Ship default `rbac_model.conf` + `rbac_policy.csv` so casbin/policy commands work OOTB.

**P1 — fix before "enterprise standard" positioning**
8. Full docs reconciliation pass per §8 (cli-reference flags, deployment-guide env/config keys, plugin guide interface, api-reference `api/` dir, testing-guide claims, truncated database-guide section); rewrite readiness-assessment honestly.
9. Implement or remove stub commands: `status`, `logs`, `check-api-spec`; finish `deploy` or mark experimental; make `plugin list` scan the right directory; `dockerize` Go version + `--tag`.
10. Tests for `framework/errors` (zero today), database, health, circuitbreaker, validation, worker; raise CI gate stepwise (28→50→80) to match the documented promise.
11. Add CONTRIBUTING.md, SECURITY.md, CODE_OF_CONDUCT.md, CHANGELOG.md; Dockerfile HEALTHCHECK; ship the k8s/helm assets the deployment guide describes (or cut those sections).

**P2 — competitive parity roadmap (already correctly identified in `docs/roadmap/`)**
12. Real Ent (or sqlc/GORM) integration; OpenAPI-first generation + published spec; `generate` emitting tests; docs website; Vault/mTLS/rate-limiting/outbox/DLQ per the enterprise-readiness roadmap; binary plugin protocol for the CLI.

---

## 11. Bottom line

Axiomod's engineering core is real and broadly well-built — 22 working framework packages, 12 of 13 working plugins, a genuinely useful validator suite, green race-enabled tests, and honest roadmap docs. What stands between this repo and an "enterprise standard" launch is not missing features but **product integrity**: the first-run experience, config correctness, self-consistent architecture rules, release engineering, and ~40 documented claims that the code does not honor. All P0 items are fixable in a focused sprint; none require new feature development.
