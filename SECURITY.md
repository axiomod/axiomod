# Security Policy

## Supported Versions

| Version | Supported |
| ------- | --------- |
| 0.2.x   | ✅        |
| < 0.2   | ❌        |

## Reporting a Vulnerability

Please **do not** open a public issue for security vulnerabilities.

Report privately via GitHub's **Security → Report a vulnerability** (private
vulnerability reporting) on this repository. Include:

- A description of the issue and its impact
- Steps to reproduce or a proof of concept
- Affected versions/commits

You will receive an acknowledgement within 72 hours. We aim to ship a fix or
mitigation within 30 days for confirmed issues and will credit reporters in
the release notes unless you prefer otherwise.

## Hardening Guidance

- Never deploy the placeholder JWT secret in `configs/service_default.yaml`;
  override it (e.g. `APP_AUTH_JWT_SECRETKEY`, minimum 32 bytes — startup
  fails otherwise).
- OIDC token verification is JWKS-signature-based and non-bypassable; set
  `auth.oidc.issuerUrl` to enable it.
- See [docs/auth-security-guide.md](docs/auth-security-guide.md) for JWT,
  OIDC, and Casbin RBAC configuration.
