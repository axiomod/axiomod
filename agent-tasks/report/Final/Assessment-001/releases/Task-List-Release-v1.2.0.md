# Task List - Release v1.2.0

# P0: Critical Priority (Must Complete Before Any Production Use)

## Epic P0.1: Security Hardening

### Feature P0.1.1: OIDC Token Verification

**Goal:** Implement full OpenID Connect token verification with signature validation, expiration checks, and audience validation.

**Rationale:** The current OIDC implementation is a critical security vulnerability. Tokens are parsed but not verified, allowing any malformed or forged token to be accepted. This is unacceptable for any production system.

**Scope:**

*In:*

- Implement JWKS (JSON Web Key Set) fetching from the OIDC issuer.

- Implement token signature verification using the fetched public keys.

- Implement token expiration validation.

- Implement token audience (`aud`) claim validation.

- Implement JWKS caching with TTL and refresh logic.

- Add error handling for key rotation and discovery failures.

- Support multiple OIDC providers (future-proofing).

*Out:*

- OIDC user provisioning or account linking (future task).

- Multi-issuer support in a single request (future task).

**Dependencies:**

- `github.com/golang-jwt/jwt/v5` (already present).

- `github.com/lestrrat-go/jwx` or similar JWKS library.

**Risks:**

- Key rotation during deployment could cause service disruption. *Mitigation:* Implement graceful fallback to cached keys with extended TTL.

- Incompatible OIDC providers. *Mitigation:* Test with multiple providers (Keycloak, Auth0, Okta).

- Performance impact of JWKS fetching. *Mitigation:* Implement aggressive caching and async refresh.

**Acceptance Criteria:**

- [ ] OIDC tokens with valid signatures are accepted.

- [ ] Tokens with invalid signatures are rejected with clear error messages.

- [ ] Expired tokens are rejected.

- [ ] Tokens with mismatched audience are rejected.

- [ ] JWKS cache is refreshed every 1 hour (configurable).

- [ ] Unit tests cover all verification paths (valid, invalid, expired, wrong audience).

- [ ] Integration test with a mock OIDC provider (e.g., Docker-based Keycloak).

- [ ] Performance test shows <50ms latency for token verification with cached JWKS.

**Test Plan:**

- Unit tests for signature verification, expiration, and audience validation.

- Integration test with a containerized Keycloak instance.

- Negative tests for malformed tokens, expired tokens, and invalid signatures.

- Performance test with 1000 concurrent token verifications.

**Documentation Needs:**

- Update `docs/auth-security-guide.md` with OIDC configuration and troubleshooting.

- Add example configuration for common OIDC providers (Keycloak, Auth0, Okta).

- Document JWKS caching behavior and TTL configuration.

- Add troubleshooting guide for common OIDC issues.

**Release Notes Entry:**

```
## Security
- **BREAKING:** OIDC token verification is now enforced. Tokens without valid signatures will be rejected. Update your OIDC provider configuration if needed.
- **NEW:** JWKS caching is implemented to reduce network calls and improve performance.
```


---

### Feature P0.1.2: RBAC Integration (Casbin)

**Goal:** Implement role-based access control (RBAC) using Casbin, with middleware for HTTP and gRPC.

**Rationale:** The framework claims RBAC support but has no integration with Casbin or any authorization library. Roles in JWT claims are purely informational without enforcement, creating a false sense of security.

**Scope:**

*In:*

- Add Casbin dependency and initialize enforcer from configuration.

- Implement Fiber middleware for RBAC enforcement.

- Implement gRPC interceptor for RBAC enforcement.

- Provide plugin for Casbin configuration.

- Support role definitions and policy files (YAML, database).

- Implement policy caching and refresh logic.

- Add support for custom matchers and functions.

*Out:*

- Dynamic policy updates at runtime (future task).

- Attribute-based access control (ABAC) (future task).

- Policy versioning and audit trails (future task).

**Dependencies:**

- `github.com/casbin/casbin/v2`.

**Risks:**

- Policy file syntax errors could break authorization. *Mitigation:* Validate policies at startup; fail fast with clear error messages.

- Performance impact of policy evaluation. *Mitigation:* Benchmark with realistic policy sets; implement caching.

- Complexity of policy definition. *Mitigation:* Provide templates and examples for common scenarios.

**Acceptance Criteria:**

- [ ] Casbin enforcer is initialized from configuration.

- [ ] HTTP requests are checked against policies before reaching handlers.

- [ ] gRPC requests are checked against policies before reaching service methods.

- [ ] Policies can be defined in YAML or database.

- [ ] Policy evaluation errors are logged and handled gracefully.

- [ ] Unit tests cover all authorization paths (allow/deny/error).

- [ ] Integration tests verify middleware/interceptor behavior.

- [ ] Performance test shows <10ms policy evaluation overhead.

**Test Plan:**

- Unit tests for policy enforcement (allow/deny scenarios, error cases).

- Integration tests with sample policies and requests.

- Performance tests to measure policy evaluation overhead.

- Test with multiple policy models (RBAC, ABAC, ACL).

**Documentation Needs:**

- Add `docs/rbac-guide.md` with policy definition examples.

- Document Casbin model and policy syntax.

- Provide example policies for common scenarios (admin, user, guest).

- Add troubleshooting guide for policy evaluation issues.

**Release Notes Entry:**

```
## Features
- **NEW:** Role-based access control (RBAC) via Casbin. Define policies in YAML or database and enforce them automatically on HTTP and gRPC endpoints.
- **NEW:** Policy caching and refresh logic for improved performance.
```



---

## Epic P0.2: Build & Version Stability

### Feature P0.2.1: Go Version Compatibility

**Goal:** Standardize on a supported Go version and ensure the framework builds reliably.

**Rationale:** The framework requires Go 1.24.2, but it should support newer version 1.25 as well.

**Scope:**

*In:*

- Evaluate minimum Go version required by dependencies.

- Update `go.mod` to specify a realistic minimum version (e.g., 1.24).

- Test build with Go 1.24, and 1.25.

- Update documentation with Go version requirements.

- Create a `.go-version` file for version management tools.

- Add CI/CD checks for multiple Go versions.

*Out:*

- Backporting to Go 1.18 or earlier.

**Dependencies:**

- None (internal task).

**Risks:**

- Dependencies may require newer Go versions. *Mitigation:* Audit all dependencies before finalizing the version.

- Breaking changes in newer Go versions. *Mitigation:* Test thoroughly with each version.

**Acceptance Criteria:**

- [ ] Framework builds successfully with Go 1.24+.

- [ ] All tests pass with Go 1.24, 1.25.

- [ ] `go.mod` specifies the minimum supported version.

- [ ] CI/CD pipeline tests against multiple Go versions.

- [ ] Documentation clearly states version requirements.

- [ ] `.go-version` file is created and documented.

**Test Plan:**

- Build tests with multiple Go versions (local and CI/CD).

- Run full test suite with each version.

- Test dependency compatibility with each version.

**Documentation Needs:**

- Update `README.md` with Go version requirements.

- Update `docs/deployment-guide.md` with installation instructions.

- Document version management tool setup.

**Release Notes Entry:**

```
## Infrastructure
- **BREAKING:** Minimum Go version is now 1.24. Upgrade your Go installation before building.
- **NEW:** CI/CD pipeline now tests against multiple Go versions (1.25, 1.25).
```


---

## Epic P0.3: Test Coverage

### Feature P0.3.1: Core Module Test Coverage

**Goal:** Achieve >80% test coverage for all critical framework modules.

**Rationale:** The framework has only 14 test files for ~14,000 lines of code. This makes it difficult to ensure stability and catch regressions. Low coverage is a major risk for a framework that will be used as the foundation for multiple services.

**Scope:**

*In:*

- Write unit tests for `framework/config`, `framework/auth`, `framework/database`, `framework/validation`, `framework/middleware`, `framework/kafka`, `framework/worker`, `framework/circuitbreaker`, `framework/health`.

- Write integration tests for server startup, shutdown, and endpoint behavior.

- Set up code coverage tracking (e.g., Codecov).

- Add coverage gates to CI/CD (fail if coverage drops below 80%).

- Document testing patterns and best practices.

*Out:*

- E2E tests (future task).

- Load/stress tests (future task).

**Dependencies:**

- `github.com/stretchr/testify` (already present).

**Risks:**

- Writing tests takes significant time. *Mitigation:* Prioritize critical paths first; use table-driven tests for efficiency.

- Difficulty achieving 80% coverage in some modules. *Mitigation:* Set realistic targets; document exceptions.

**Acceptance Criteria:**

- [ ] >80% coverage for all core modules.

- [ ] All tests pass consistently.

- [ ] Coverage is tracked and reported in CI/CD.

- [ ] Coverage gates prevent regressions.

- [ ] Testing patterns are documented.

**Test Plan:**

- Write unit tests for each module.

- Write integration tests for server lifecycle.

- Run coverage reports and identify gaps.

- Establish baseline and track over time.

**Documentation Needs:**

- Add `docs/testing-guide.md` with testing patterns and best practices.

- Document table-driven test patterns.

- Provide examples of unit and integration tests.

**Release Notes Entry:**

```
## Quality
- **NEW:** Test coverage for core modules now exceeds 80%. See testing-guide.md for patterns.
- **NEW:** CI/CD pipeline enforces minimum coverage gates.
```

