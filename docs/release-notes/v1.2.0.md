# v1.2.0 - Security & Stability Hardening 🛡️

This release focuses on strengthening the security foundations of the framework and Improving overall build stability and test coverage.

## 🌟 Key Highlights

### 🔐 Security & Authorization

- **OIDC Signature Verification**: Mandatory JWKS-based token verification is now enforced. Tokens are verified against the provider's public keys using OIDC discovery.
- **RBAC Integration via Casbin**: Role-Based Access Control is now integrated using Casbin.
  - Standard `RBACService` for programmatic checks.
  - Fiber middleware for route-level enforcement.
  - gRPC interceptor for service-level enforcement.
- **JWKS Caching**: Automated retrieval and caching of JSON Web Key Sets to minimize latency and improve performance.

### 🏗️ Build & Stability

- **Go Version Standardization**: Standardized on Go 1.24.2+. Added `.go-version` support.
- **CI/CD Enhancements**: Added support for multi-version Go testing (1.24, 1.25).

### 🧪 Quality & Testing

- **Core Module Coverage**: Significantly increased test coverage across all core modules to >80%.
- **New Test Suites**:
  - `framework/config`
  - `framework/auth` (JWT, OIDC, RBAC)
  - `framework/validation`
  - `framework/middleware`
  - `framework/health`
  - `framework/circuitbreaker`
- **Testing Patterns**: Introduced standardized table-driven testing patterns and a new [Testing Guide](./docs/testing-guide.md).

## ⚠️ Breaking Changes

- **Minimum Go Version**: The framework now requires Go 1.24 or higher.
- **OIDC Verification**: `OIDCService.VerifyToken` now strictly enforces signature verification. Ensure your OIDC provider is correctly configured for discovery.

## 📚 New Documentation

- [Auth & Security Guide Update](./docs/auth-security-guide.md): Detailed OIDC and RBAC configuration.
- [Testing Guide](./docs/testing-guide.md): Best practices and patterns.
- [ADR-003: OIDC Signature Verification](./docs/decision-records/ADR-003-oidc-signature-verification.md)
- [ADR-004: RBAC Casbin Integration](./docs/decision-records/ADR-004-rbac-casbin-integration.md)
