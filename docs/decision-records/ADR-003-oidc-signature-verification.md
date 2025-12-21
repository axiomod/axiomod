# ADR-003: Mandatory OIDC Signature Verification via JWKS

## Status

Accepted

## Context

The framework previously lacked robust OIDC token verification. Specifically, it did not enforce signature validation, which is a critical security requirement for OIDC. Without signature verification, the framework could be vulnerable to token spoofing and tampering.

## Decision

We have implemented mandatory signature verification for OIDC tokens using JSON Web Key Sets (JWKS).

- Integrated `github.com/MicahParks/keyfunc/v3` to handle JWKS retrieval and management.
- The `OIDCService` now automatically performs OIDC discovery to locate the JWKS endpoint.
- Public keys from the JWKS endpoint are cached locally to minimize network latency and reduce load on the OIDC provider.
- Signature verification is integrated into the `VerifyToken` method and cannot be bypassed.

## Consequences

- **Security**: Significantly improved security posture by ensuring token authenticity and integrity.
- **Performance**: JWKS caching ensures minimal latency impact (< 50ms) for subsequent verifications.
- **Reliability**: Automatic discovery and JWKS management reduce manual configuration errors.
- **Dependency**: Introduced a new dependency on `github.com/MicahParks/keyfunc/v3`.
