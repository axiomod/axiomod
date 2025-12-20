# Plugin Development Guide

## Introduction

The Enterprise Axiomod provides a flexible plugin system that allows you to extend the framework's functionality without modifying its core. This guide will walk you through the process of creating, configuring, and using plugins.

## Plugin Interface

All plugins must implement the `Plugin` interface defined in `plugins/plugin.go`:

```go
type Plugin interface {
    // Name returns the name of the plugin
    Name() string
    
    // Initialize initializes the plugin with the given configuration and logger
    Initialize(config map[string]interface{}, logger *zap.Logger) error
    
    // Start starts the plugin
    Start() error
    
    // Stop stops the plugin
    Stop() error
}
```

## Creating a New Plugin

### 1. Create a new package

Create a new package for your plugin in the `plugins` directory:

```bash
mkdir -p plugins/my_plugin
```

### 2. Implement the Plugin interface

Create a new file `my_plugin.go` in the package and implement the Plugin interface:

```go
package my_plugin

import (
    "github.com/axiomod/axiomod/plugins"
)

// MyPlugin is a custom plugin implementation
type MyPlugin struct {
    config map[string]interface{}
    active bool
}

// Name returns the name of the plugin
func (p *MyPlugin) Name() string {
    return "my_plugin"
}

// Initialize initializes the plugin with the given configuration
func (p *MyPlugin) Initialize(config map[string]interface{}, logger *zap.Logger) error {
    p.config = config
    return nil
}

// Start starts the plugin
func (p *MyPlugin) Start() error {
    p.active = true
    return nil
}

// Stop stops the plugin
func (p *MyPlugin) Stop() error {
    p.active = false
    return nil
}

// IsActive returns whether the plugin is active
func (p *MyPlugin) IsActive() bool {
    return p.active
}
```

### 3. Register the plugin

Register your plugin in the `registerBuiltInPlugins` method in `plugins/plugin.go`:

```go
func (r *PluginRegistry) registerBuiltInPlugins() {
    // Existing plugins...
    
    // Register your plugin
    r.Register(&my_plugin.MyPlugin{})
}
```

## Plugin Configuration

Plugins are configured through the central configuration system. You can configure your plugin in the `config.yaml` file:

```yaml
plugins:
  enabled:
    - my_plugin
  config:
    my_plugin:
      option1: value1
      option2: value2
```

Or through environment variables:

```bash
PLUGINS_ENABLED=my_plugin
MY_PLUGIN_OPTION1=value1
MY_PLUGIN_OPTION2=value2
```

## Plugin Types

The framework supports several types of plugins:

### Database Plugins

Database plugins provide access to different database systems. They should implement the `Plugin` interface and provide a way to get a database connection.

Example:

```go
type DatabasePlugin interface {
    plugins.Plugin
    GetConnection() (*sql.DB, error)
}
```

### Authentication Plugins

Authentication plugins provide authentication mechanisms. They should implement the `Plugin` interface and provide methods for authentication and authorization.

Example:

```go
type AuthPlugin interface {
    plugins.Plugin
    Authenticate(token string) (User, error)
    Authorize(user User, resource string, action string) bool
}
```

### Feature Flag Plugins

Feature flag plugins provide feature flag functionality. They should implement the `Plugin` interface and provide methods for checking if a feature is enabled.

Example:

```go
type FeatureFlagPlugin interface {
    plugins.Plugin
    IsEnabled(feature string) bool
}
```

## Plugin Lifecycle

Plugins go through the following lifecycle:

1. **Registration**: Plugins are registered with the plugin registry
2. **Initialization**: Plugins are initialized with their configuration
3. **Start**: Plugins are started when the application starts
4. **Stop**: Plugins are stopped when the application stops

## Best Practices

### 1. Keep plugins focused

Each plugin should have a single responsibility. If you find your plugin doing too many things, consider splitting it into multiple plugins.

### 2. Handle errors gracefully

Plugins should handle errors gracefully and not panic. If a plugin encounters an error, it should log the error and return it.

### 3. Use dependency injection

Plugins should use dependency injection to get their dependencies. This makes them more testable and maintainable.

### 4. Document your plugin

Provide documentation for your plugin, including its purpose, configuration options, and usage examples.

### 5. Write tests

Write tests for your plugin to ensure it works as expected. Use mocks for external dependencies.

## Example: Creating a Simple Greeter Plugin

Here's an example of creating a simple plugin that logs a greeting:

```go
package greeter

import (
    "fmt"
    "github.com/axiomod/axiomod/plugins"
    "go.uber.org/zap"
)

// GreeterPlugin implements a greeting plugin
type GreeterPlugin struct {
    config map[string]interface{}
    logger *zap.Logger
}

// Name returns the name of the plugin
func (p *GreeterPlugin) Name() string {
    return "greeter"
}

// Initialize initializes the plugin configuration
func (p *GreeterPlugin) Initialize(config map[string]interface{}, logger *zap.Logger) error {
    p.config = config
    p.logger = logger
    return nil
}

// Start starts the plugin
func (p *GreeterPlugin) Start() error {
    greeting, _ := p.config["greeting"].(string)
    if greeting == "" {
        greeting = "Hello"
    }
    p.logger.Info(fmt.Sprintf("%s from GreeterPlugin!", greeting))
    return nil
}

// Stop stops the plugin
func (p *GreeterPlugin) Stop() error {
    p.logger.Info("GreeterPlugin stopped")
    return nil
}
```

## Conclusion

The plugin system provides a flexible way to extend the framework's functionality. By following this guide, you can create your own plugins to add new features to the framework.
