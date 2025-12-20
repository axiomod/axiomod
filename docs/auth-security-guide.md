# Authentication & Security Guide

This guide covers the authentication mechanisms provided by the Axiomod framework, including JWT and OIDC (Keycloak) integration.

## 1. JWT Authentication

The framework provides a built-in JWT service for token generation and verification.

### Configuration

Update the `auth` section in your configuration:

```yaml
auth:
  provider: jwt
  jwtSecret: your-long-and-secure-secret
  jwtDuration: 3600  # seconds (1 hour)
```

### Usage

Inject the `auth.JWTService` into your handlers or services:

```go
func (s *LoginService) Login(ctx context.Context, user, pass string) (string, error) {
    // ... validate user credentials ...
    
    // Generate token
    return s.jwtService.GenerateToken(userID, username, email, roles)
}
```

### Middleware

The `internal/framework/middleware` package provides an `AuthMiddleware` (for Fiber) that automatically validates incoming JWT tokens in the `Authorization: Bearer <token>` header.

## 2. OIDC / Keycloak Integration

For enterprise environments, the framework supports OIDC discovery and token verification.

### Configuration

Settings for Keycloak are managed via the `KeycloakPlugin`:

```yaml
plugins:
  enabled:
    - keycloak
  settings:
    keycloak:
      issuer: "https://keycloak.example.com/realms/master"
      client_id: "axiomod-client"
      client_secret: "..."
```

### OIDC Discovery

When the `KeycloakPlugin` starts, it automatically performs OIDC discovery to fetch the authorization, token, and JWKS endpoints.

### Usage in Code

The `OIDCService` in `internal/framework/auth/oidc.go` can be used to verify OIDC tokens:

```go
claims, err := oidcService.VerifyToken(ctx, tokenString)
if err != nil {
    return nil, errors.ErrUnauthorized
}
```

## 3. Best Practices

### Secret Management
>
> [!IMPORTANT]
> Never commit secrets (JWT secrets, DB passwords, OIDC client secrets) to your version control system.

- Use **Environment Variables** in production.
- Integrations with secret managers like **HashiCorp Vault** are recommended for high-security environments.

### Token Expiration

- Keep JWT durations short (e.g., 1 hour).
- Use refresh tokens if long-lived sessions are required.

### RBAC (Role-Based Access Control)

Roles are included in the JWT/OIDC claims. Use the `claims.HasRole("admin")` helper to check for specific permissions in your use cases.
