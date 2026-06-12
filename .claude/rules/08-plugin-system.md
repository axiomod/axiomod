# Plugin System

## Plugin Interface

All plugins implement the interface in `plugins/plugin.go`:

```go
type Plugin interface {
    Name() string
    Initialize(config map[string]interface{}, logger *observability.Logger,
               metrics *observability.Metrics, cfg *config.Config, health *health.Health) error
    Start() error
    Stop() error
}
```

## Plugin Implementation Pattern

```go
type MyPlugin struct {
    logger  *observability.Logger
    metrics *observability.Metrics
    config  map[string]interface{}
}

func (p *MyPlugin) Name() string { return "my-plugin" }

func (p *MyPlugin) Initialize(config map[string]interface{}, logger *observability.Logger,
    metrics *observability.Metrics, cfg *config.Config, health *health.Health) error {
    p.config = config
    p.logger = logger
    p.metrics = metrics
    // Register health check
    health.RegisterCheck(p.Name(), func() error { return p.ping() })
    return nil
}

func (p *MyPlugin) Start() error {
    // Start connections, background workers, etc.
    return nil
}

func (p *MyPlugin) Stop() error {
    // Clean shutdown
    return nil
}
```

## Registration

Register via `PluginRegistry.Register(&MyPlugin{})` in the `RegisterNewPlugins` function.

## Built-in Plugins

| Plugin | Config key | Type | Package |
|---|---|---|---|
| `MySQLPlugin` | `mysql` | Database | `plugins/` |
| `PostgreSQLPlugin` | `postgres` | Database | `plugins/` |
| `JWTPlugin` | `jwt` | Auth | `plugins/` |
| `KeycloakPlugin` | `keycloak` | Auth | `plugins/` |
| `CasdoorPlugin` | `casdoor` | Auth | `plugins/` |
| `CasbinPlugin` | `casbin` | RBAC | `plugins/` |

Extended (registered via `RegisterNewPlugins` in `cmd/axiomod-server`):

| Plugin | Config key | Package |
|---|---|---|
| LDAP | `ldap` | `plugins/auth/ldap` |
| SAML | `saml` | `plugins/auth/saml` |
| Multitenancy | `multitenancy` | `plugins/middleware/multitenancy` |
| Audit | `auditing` | `plugins/audit` |
| ELK | `elk` | `plugins/logging/elk` |
| Redis | `redis` | `plugins/cache/redis` |
| Kafka | `kafka` | `plugins/messaging/kafka` |

## Configuration

Enable/disable plugins in `configs/service_default.yaml`:

```yaml
plugins:
  enabled:
    postgres: true
    mysql: false
    jwt: true
  settings:
    multitenancy:
      header: "X-Tenant-ID"
```

Maps to `framework/config/plugins_config.go`:

```go
type PluginsConfig struct {
    Enabled  map[string]bool
    Settings map[string]map[string]interface{}
    Paths    []string
}
```

## Plugin Lifecycle

1. `Register()` -- Adds plugin to registry (built-ins at construction,
   extended plugins via `fx.Invoke(RegisterNewPlugins)`)
2. `Initialize()` -- Called for enabled plugins with their config inside
   `StartAll()` (fx `OnStart`), after all registrations
3. `Start()` -- Called right after `Initialize()` in `StartAll()`
4. `Stop()` -- Called via fx `OnStop` hook

`StartAll()` fails fast if a plugin enabled in `plugins.enabled` is not
registered, or if any `Initialize()`/`Start()` returns an error.

## Rules

1. Every plugin MUST implement all 4 interface methods
2. `Name()` must return a unique, lowercase, hyphenated string
3. Register health checks in `Initialize()`
4. Handle graceful shutdown in `Stop()`
5. Plugins may import `platform/*` and `framework/*` only
