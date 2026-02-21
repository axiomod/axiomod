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

| Plugin | Type | Package |
|---|---|---|
| `MySQLPlugin` | Database | `plugins/` |
| `PostgreSQLPlugin` | Database | `plugins/` |
| `JWTPlugin` | Auth | `plugins/` |
| `KeycloakPlugin` | Auth | `plugins/` |
| `CasdoorPlugin` | Auth | `plugins/` |
| `CasbinPlugin` | RBAC | `plugins/` |

Extended: `ldap`, `saml`, `multitenancy`, `audit`, `elk` (in subdirectories).

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

1. `Register()` -- Adds plugin to registry
2. `Initialize()` -- Called for enabled plugins with their config
3. `Start()` -- Called via fx `OnStart` hook
4. `Stop()` -- Called via fx `OnStop` hook (reverse order)

## Rules

1. Every plugin MUST implement all 4 interface methods
2. `Name()` must return a unique, lowercase, hyphenated string
3. Register health checks in `Initialize()`
4. Handle graceful shutdown in `Stop()`
5. Plugins may import `platform/*` and `framework/*` only
