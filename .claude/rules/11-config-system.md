# Configuration System

## Config Structure

Central config in `framework/config/types.go`:

```go
type Config struct {
    App           AppConfig
    Observability ObservabilityConfig
    Database      DatabaseConfig
    HTTP          HTTPConfig
    GRPC          GRPCConfig
    Auth          AuthConfig
    Casbin        CasbinConfig
    Plugins       PluginsConfig
}
```

## Config Naming Conventions

### YAML Keys

**Convention: camelCase** throughout `configs/service_default.yaml`:

```yaml
app:
  name: "axiomod"
  environment: "development"

observability:
  logLevel: "info"            # camelCase
  logFormat: "json"
  tracingEnabled: true
  tracingExporterType: "jaeger"
  tracingSamplerRatio: 1.0
  metricsEnabled: true
  metricsPort: 9100

database:
  sslMode: "disable"
  maxOpenConns: 25
  maxIdleConns: 10
  connMaxLifetime: "5m"
  slowQueryThreshold: "200ms"

auth:
  oidc:
    issuerUrl: "http://..."   # Note: lowercase 'rl' in YAML (not issuerURL)
    clientId: "my-client"     # Note: lowercase 'd' in YAML (not clientID)
    jwksCacheTtl: "1h"       # Note: lowercase 'ttl' in YAML
```

**Top-level keys**: Single lowercase words (`app`, `observability`, `database`, `http`, `grpc`, `auth`, `casbin`, `plugins`)

**Nested keys**: camelCase. Acronyms are lowered in YAML (`issuerUrl`, `clientId`, `jwksCacheTtl`, `sslMode`).

### Go Config Struct Fields

**Convention: PascalCase** with fully capitalized acronyms:

```go
type AuthConfig struct {
    OIDC OIDCConfig
    JWT  JWTConfig
}

type OIDCConfig struct {
    IssuerURL    string    // YAML: issuerUrl (Viper maps case-insensitively)
    ClientID     string    // YAML: clientId
    ClientSecret string    // YAML: clientSecret
    JWKSCacheTTL string   // YAML: jwksCacheTtl
}
```

**No struct tags**: Config structs have NO `mapstructure:`, `yaml:`, or `env:` tags. Viper handles case-insensitive mapping automatically.

### Environment Variables

**Prefix**: `APP_`
**Separator**: Underscore replaces dots (`.` -> `_`)
**Case**: Viper's auto-env is case-insensitive

Mapping: Viper key `section.field` -> env var `APP_SECTION_FIELD`

```
app.name                    -> APP_APP_NAME
http.port                   -> APP_HTTP_PORT
observability.logLevel      -> APP_OBSERVABILITY_LOGLEVEL
observability.tracingEnabled -> APP_OBSERVABILITY_TRACINGENABLED
database.maxOpenConns       -> APP_DATABASE_MAXOPENCONNS
```

**Note**: No camelCase-to-SCREAMING_SNAKE conversion exists. The env var key preserves the camelCase part joined: `LOGLEVEL` not `LOG_LEVEL`.

### Viper Config Keys

Dot-separated, lowercase: `app.name`, `http.port`, `database.host`, `plugins.enabled`

### Plugin Names in Config

**Lowercase** single words in the `plugins.enabled` map:

```yaml
plugins:
  enabled:
    postgres: true
    mysql: false
    jwt: true
    keycloak: false
    multitenancy: true
  settings:
    multitenancy:
      header: "X-Tenant-ID"
```

## Viper-Based Loading

- Reads `configs/service_default.yaml` by default
- Search paths: `.`, `config`, `../config`, `../../config`, `framework/config`
- Supports config file watching via `WatchConfig(onChange func())`

## Provider Interface

```go
type Provider interface {
    Get(key string) interface{}
    GetString(key string) string
    GetInt(key string) int
    GetBool(key string) bool
    GetFloat64(key string) float64
    GetDuration(key string) time.Duration
    GetStringSlice(key string) []string
    GetStringMap(key string) map[string]interface{}
    IsSet(key string) bool
    AllSettings() map[string]interface{}
    WatchConfig(onChange func())
}
```

## Specialized Loaders

```go
LoadServiceConfig()      // configs/service_default.yaml
LoadPluginConfig()       // plugin_settings.yaml
LoadCLIConfig()          // cli_config.yaml
LoadConfigFile(path)     // arbitrary file
```

## Default Config Functions

Packages provide their own defaults:

```go
tracing.DefaultConfig()
kafka.DefaultProducerConfig()
kafka.DefaultConsumerConfig()
circuitbreaker.DefaultOptions()
resilience.DefaultRetryOptions()
```

## Rules

1. All config through `*config.Config` -- never hardcode values
2. New config sections: add struct to `types.go` with `<Domain>Config` naming
3. YAML keys: always **camelCase** (lowercase acronyms: `issuerUrl` not `issuerURL`)
4. Go struct fields: always **PascalCase** (uppercase acronyms: `IssuerURL` not `IssuerUrl`)
5. No struct tags on config types -- rely on Viper's case-insensitive matching
6. Environment override: prefix with `APP_`, dots become underscores
7. Plugin names in config: **lowercase** single words
8. Config files go in `configs/` directory
9. Use `Provider` interface for testability
10. Per-package defaults via `Default<Name>()` functions, not in central config
