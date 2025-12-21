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

The `github.com/axiomod/axiomod/framework/middleware` package provides an `AuthMiddleware` (for Fiber) that automatically validates incoming JWT tokens in the `Authorization: Bearer <token>` header.

## 2. OIDC / Keycloak Integration

For enterprise environments, the framework supports OIDC discovery and token verification.

### Provider Configuration

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

The `OIDCService` in `framework/auth/oidc.go` provides secure, JWKS-based token verification:

```go
claims, err := oidcService.VerifyToken(ctx, tokenString)
if err != nil {
    return nil, errors.ErrUnauthorized
}
```

> [!NOTE]
> Signature verification is MANDATORY. The service automatically fetches public keys from the provider's JWKS endpoint and caches them locally for 1 hour (configurable).

## 3. RBAC (Role-Based Access Control)

The framework uses **Casbin** for robust, policy-based authorization.

### Configuration

Add the `casbin` section to your configuration:

```yaml
casbin:
  modelPath: "path/to/rbac_model.conf"
  policyPath: "path/to/policy.csv"
```

### Middleware (Fiber)

Enforce permissions on HTTP routes using the `RBACMiddleware`:

```go
app.Get("/protected", 
    rbac.Handle("resource_name", "read"), 
    handler
)
```

### Interceptor (gRPC)

For gRPC services, use the `RBACInterceptor` to protect your methods:

```go
s := grpc.NewServer(
    grpc.UnaryInterceptor(
        grpc_middleware.ChainUnaryServer(
            authInterceptor,
            grpc_auth.RBACInterceptor(rbacService, logger),
        ),
    ),
)
```

### Policy Management via CLI

You can manage Casbin policies and roles directly using the `axiomod policy` command group.

```bash
# List all policies
axiomod policy list

# Add a policy
axiomod policy add --ptype=p --v0=role:admin --v1=resource --v2=action

# Remove a policy
axiomod policy remove --ptype=p --v0=role:admin --v1=resource --v2=action
```

## 4. Best Practices

### Secret Management
>
> [!IMPORTANT]
> Never commit secrets (JWT secrets, DB passwords, OIDC client secrets) to your version control system.

- Use **Environment Variables** in production.
- Integrations with secret managers like **HashiCorp Vault** are recommended for high-security environments.

### Token Expiration

- Keep JWT durations short (e.g., 1 hour).
- Use refresh tokens if long-lived sessions are required.

### Role Checks in Claims

While Casbin provides policy-based authorization, you can also perform manual role checks using the `claims.HasRole("admin")` helper on the decoded JWT/OIDC claims.
