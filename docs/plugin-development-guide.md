# Plugin Development Guide

This guide explains how to create plugins for the Axiomod framework. The
reference implementation lives at
[`plugins/example_plugin/example_plugin.go`](../plugins/example_plugin/example_plugin.go)
and compiles against the real interface — copy it as your starting point.

## The Plugin Interface

Every plugin implements the interface defined in
[`plugins/plugin.go`](../plugins/plugin.go):

```go
type Plugin interface {
    Name() string
    Initialize(config map[string]interface{}, logger *observability.Logger,
        metrics *observability.Metrics, cfg *config.Config, health *health.Health) error
    Start() error
    Stop() error
}
```

- `Name()` returns the unique, lowercase plugin name. It is the key under
  `plugins.enabled` and `plugins.settings` in the service configuration.
- `Initialize(...)` receives the plugin's own settings block
  (`plugins.settings.<name>`), the shared logger and metrics, the full
  application `*config.Config`, and the health registry. Register health
  checks here; do not open external connections yet.
- `Start()` is called after **all** enabled plugins are initialized. Open
  connections and launch background workers here. Returning an error aborts
  startup (fail fast).
- `Stop()` is called on shutdown via the fx `OnStop` hook. Release everything
  acquired in `Start()`.

## Walkthrough: the example plugin

```go
package example_plugin

import (
    "fmt"

    "github.com/axiomod/axiomod/framework/config"
    "github.com/axiomod/axiomod/framework/health"
    "github.com/axiomod/axiomod/framework/observability"

    "go.uber.org/zap"
)

type ExamplePlugin struct {
    settings map[string]interface{}
    logger   *observability.Logger
    metrics  *observability.Metrics
    greeting string
    active   bool
}

func (p *ExamplePlugin) Name() string { return "example" }

func (p *ExamplePlugin) Initialize(
    settings map[string]interface{},
    logger *observability.Logger,
    metrics *observability.Metrics,
    cfg *config.Config,
    h *health.Health,
) error {
    p.settings = settings
    p.logger = logger
    p.metrics = metrics

    p.greeting = "Hello"
    if g, ok := settings["greeting"].(string); ok && g != "" {
        p.greeting = g
    }

    h.RegisterCheck(p.Name(), func() error {
        if !p.active {
            return fmt.Errorf("example plugin is not running")
        }
        return nil
    })
    return nil
}

func (p *ExamplePlugin) Start() error {
    p.active = true
    p.logger.Info("Example plugin started", zap.String("greeting", p.greeting))
    return nil
}

func (p *ExamplePlugin) Stop() error {
    p.active = false
    p.logger.Info("Example plugin stopped")
    return nil
}
```

## Registration

There are two registration points:

1. **Built-in plugins** (database, JWT, Keycloak, Casdoor, Casbin) are
   registered in `registerBuiltInPlugins` inside
   [`plugins/builtin_plugins.go`](../plugins/builtin_plugins.go).
2. **Standalone plugins** (everything under `plugins/<category>/<name>/`)
   are registered in `RegisterNewPlugins` in
   [`cmd/axiomod-server/register_plugins.go`](../cmd/axiomod-server/register_plugins.go):

```go
func RegisterNewPlugins(r *plugins.PluginRegistry) {
    r.Register(&ldap.Plugin{})
    r.Register(&myplugin.Plugin{})  // <- add yours here
    // ...
}
```

## Configuration

Plugins are enabled per service in `configs/service_default.yaml`.
`plugins.enabled` is a **map of plugin name to boolean**, and per-plugin
settings live under `plugins.settings.<name>`:

```yaml
plugins:
  enabled:
    example: true
    redis: false
  settings:
    example:
      greeting: "Hello from config"
```

Environment overrides use the standard `APP_` prefix with Viper key dots
replaced by underscores (overrides apply to keys present in the config
file):

```bash
export APP_PLUGINS_ENABLED_EXAMPLE=true
export APP_PLUGINS_SETTINGS_EXAMPLE_GREETING="Hi"
```

## Plugin Lifecycle

1. **Register** — built-ins at registry construction; standalone plugins via
   `fx.Invoke(RegisterNewPlugins)`.
2. **Initialize** — called for enabled plugins with their settings inside
   `StartAll()` (fx `OnStart`), after all registrations.
3. **Start** — called right after `Initialize()` in `StartAll()`.
4. **Stop** — called via the fx `OnStop` hook.

`StartAll()` fails fast if a plugin enabled in `plugins.enabled` is not
registered, or if any `Initialize()`/`Start()` returns an error.

## Built-in and Standalone Plugins

| Plugin | Config key | Package |
|---|---|---|
| MySQL / PostgreSQL | `mysql` / `postgres` | `plugins/` (built-in) |
| JWT / Keycloak / Casdoor / Casbin | `jwt` / `keycloak` / `casdoor` / `casbin` | `plugins/` (built-in) |
| LDAP | `ldap` | `plugins/auth/ldap` |
| SAML | `saml` | `plugins/auth/saml` |
| Multitenancy | `multitenancy` | `plugins/middleware/multitenancy` |
| Audit | `auditing` | `plugins/audit` |
| ELK | `elk` | `plugins/logging/elk` |
| Redis | `redis` | `plugins/cache/redis` |
| Kafka | `kafka` | `plugins/messaging/kafka` |

Run `axiomod plugin list` to see the standalone plugins in your checkout.

## Best Practices

1. **Single responsibility** — one concern per plugin.
2. **Register a health check** in `Initialize()` for every external
   dependency.
3. **Fail fast** — return errors from `Start()` instead of degrading
   silently.
4. **Clean shutdown** — `Stop()` must release connections, flush buffers,
   and stop goroutines (see the ELK plugin for a buffered example).
5. **Guard against signature drift** with a compile-time assertion
   (`var _ plugins.Plugin = (*MyPlugin)(nil)`) — see
   `plugins/example_plugin/example_plugin_test.go`.
