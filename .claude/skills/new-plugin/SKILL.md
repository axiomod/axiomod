---
description: "Create a new Axiomod plugin implementing the Plugin interface. Use when the user says 'create plugin', 'new plugin', 'add plugin <name>'."
---

# Create New Plugin

## Inputs
- Plugin name (required): lowercase, single word
- Config keys needed (optional)

## Steps

1. **Create plugin file** implementing the Plugin interface:
   - `Name()` returns the config key
   - `Initialize()` stores deps + registers health check
   - `Start()` initializes connections/services
   - `Stop()` cleans up resources

   For built-in plugins, add struct to `plugins/builtin_plugins.go`.
   For external plugins, create `plugins/<name>/plugin.go`.

2. **Create tests** (`plugin_test.go`) covering:
   - `Name()` returns correct string
   - `Initialize()` succeeds with valid config
   - `Start()`/`Stop()` lifecycle
   - Health check integration

3. **Register the plugin**:
   - Built-in: add `r.Register(&<Name>Plugin{})` in `plugins/plugin.go:registerBuiltInPlugins()`
   - External: add in `cmd/axiomod-server/register_plugins.go`

4. **Add config to `configs/service_default.yaml`**:
   ```yaml
   plugins:
     enabled:
       <name>: false
     settings:
       <name>:
         # plugin-specific settings
   ```

5. **Verify**:
   ```bash
   go build ./...
   go test ./plugins/...
   ```

## Reference Plugins

- Simple stub: `plugins/builtin_plugins.go` CasdoorPlugin (minimal Start/Stop)
- Database: `plugins/builtin_plugins.go` MySQLPlugin (connects in Start, closes in Stop)
- Auth: `plugins/builtin_plugins.go` KeycloakPlugin (background OIDC discovery)
