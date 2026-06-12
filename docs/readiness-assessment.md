# Axiomod Framework: Readiness Assessment

**Date**: 2026-06-12
**Status**: **Release Candidate** — pending the first tagged release

This assessment lists only claims an evaluator can verify by running the
command shown. Methodology and the full findings history:
[`docs/audit/2026-06-12-launch-readiness/`](audit/2026-06-12-launch-readiness/feature-implementation-audit.md).

## 1. Verified Results

| Check | Command | Result |
| :--- | :--- | :--- |
| Build | `go build ./...` | ✅ PASS |
| Vet / format | `go vet ./...` · `gofmt -l .` | ✅ PASS / clean |
| Tests (race) | `go test -race ./...` | ✅ all packages green |
| Coverage | `go test -coverprofile=... ./...` | ✅ ~60% total (CI gate: 50%) |
| Architecture | `make validate-arch` | ✅ 0 violations (enforced in CI) |
| Server boot | `go run ./cmd/axiomod-server` | ✅ HTTP :8080, gRPC :9090, graceful shutdown, no error logs |
| Quick start | `axiomod init x --dev` → `go mod tidy && go build ./...` | ✅ compiles (CI smoke test) |
| Example API | `go test -run TestExampleModuleCRUD ./cmd/axiomod-server/` | ✅ full CRUD over HTTP with JWT auth |

## 2. Feature Readiness

| Feature Area | Status | Evidence |
| :--- | :--- | :--- |
| **Architecture** | 🟢 Ready | Layering enforced by `validator architecture` in CI; framework/platform direction consistent |
| **Database** | 🟢 Ready | postgres + mysql drivers registered; Ent default ORM with `database.orm` switch; migrations for both drivers |
| **Auth** | 🟢 Ready | JWT (≥32-byte secret enforced), OIDC with JWKS signature verification (opt-in via issuerUrl), Casbin RBAC with shipped model/policy files |
| **Observability** | 🟢 Ready | zap logging, Prometheus at `/metrics` (HTTP port), OTel tracing (jaeger/otlp/stdout) |
| **Async** | 🟢 Ready | Kafka producer/consumer (Sarama), worker pool |
| **Plugins** | 🟢 Ready | 13 plugins; registry fail-fast; settings keys normalized (camelCase YAML works) |
| **Example domain** | 🟢 Ready | Registered in the server; full CRUD over HTTP + gRPC; integration-tested |
| **CLI** | 🟡 Beta | All commands functional or honestly labeled (`deploy` defaults to dry-run); generators emit gofmt-clean code |
| **Releases** | 🔴 Pending | GoReleaser + workflow ready; **no tag published yet** — `go get` requires the first release (maintainer action) |

## 3. Known Gaps (tracked)

1. **First release not cut** — `docs/release-checklist.md` flow is ready;
   tagging is a maintainer action.
2. **OpenAPI generation** not built-in (validator exists; generation is on
   the [CLI roadmap](roadmap/cli-enhancement.md)).
3. **Coverage** at ~60% against the 80% target for core modules; CI gate
   ratchets upward.
4. Roadmap items (Vault, mTLS, rate limiting, outbox/DLQ, monorepo tooling)
   remain planned — see [docs/roadmap.md](roadmap.md).

## 4. Recommendation

Cut `v0.3.0` via the release checklist, verify
`go get github.com/axiomod/axiomod@v0.3.0` from a scratch module, then
re-run this assessment and flip **Releases** to 🟢.
