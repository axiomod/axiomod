package plugins

import (
	"context"
	"fmt"
	"sync"

	"axiomod/internal/framework/config"
	"axiomod/internal/platform/observability"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Module provides the fx options for the plugins module
var Module = fx.Options(
	fx.Provide(NewPluginRegistry),
	fx.Invoke(RegisterPlugins),
)

// RegisterPlugins registers the plugin registry with the fx lifecycle
func RegisterPlugins(lc fx.Lifecycle, registry *PluginRegistry) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return registry.StartAll()
		},
		OnStop: func(ctx context.Context) error {
			return registry.StopAll()
		},
	})
}

// Plugin defines the interface that all plugins must implement
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

// PluginRegistry manages the registration and lifecycle of plugins
type PluginRegistry struct {
	plugins map[string]Plugin
	config  *config.Config
	logger  *observability.Logger
	mu      sync.RWMutex
}

// NewPluginRegistry creates a new plugin registry
func NewPluginRegistry(cfg *config.Config, logger *observability.Logger) (*PluginRegistry, error) {
	registry := &PluginRegistry{
		plugins: make(map[string]Plugin),
		config:  cfg,
		logger:  logger,
	}

	// Register built-in plugins
	registry.registerBuiltInPlugins()

	// Initialize enabled plugins
	if err := registry.initializeEnabledPlugins(); err != nil {
		return nil, err
	}

	return registry, nil
}

// registerBuiltInPlugins registers all built-in plugins
func (r *PluginRegistry) registerBuiltInPlugins() {
	// Register database plugins
	r.Register(&MySQLPlugin{})
	r.Register(&PostgreSQLPlugin{})

	// Register auth plugins
	r.Register(&JWTPlugin{})
	r.Register(&KeycloakPlugin{})
	r.Register(&CasdoorPlugin{})

	// Register other plugins
	r.Register(&CasbinPlugin{})
	r.Register(&LDAPPlugin{})
	r.Register(&SAMLPlugin{})
	r.Register(&MultiTenancyPlugin{})
	r.Register(&AuditingPlugin{})
	r.Register(&ELKPlugin{})
}

// Register registers a plugin with the registry
func (r *PluginRegistry) Register(plugin Plugin) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.plugins[plugin.Name()] = plugin
	r.logger.Info("Registered plugin", zap.String("name", plugin.Name()))
}

// Get returns a plugin by name
func (r *PluginRegistry) Get(name string) (Plugin, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plugin, ok := r.plugins[name]
	if !ok {
		return nil, fmt.Errorf("plugin not found: %s", name)
	}

	return plugin, nil
}

// initializeEnabledPlugins initializes all enabled plugins
func (r *PluginRegistry) initializeEnabledPlugins() error {
	// Iterate over the map of enabled plugins
	for name, enabled := range r.config.Plugins.Enabled {
		if !enabled {
			continue // Skip disabled plugins
		}

		plugin, err := r.Get(name)
		if err != nil {
			// Log error but continue, maybe plugin wasn't registered
			r.logger.Error("Plugin defined in config but not found in registry", zap.String("name", name), zap.Error(err))
			continue
		}

		// Get plugin settings
		pluginSettings, ok := r.config.Plugins.Settings[name]
		if !ok {
			pluginSettings = make(map[string]interface{}) // Use empty settings if none found
		}

		// Initialize plugin
		if err := plugin.Initialize(pluginSettings, r.logger.Logger); err != nil {
			return fmt.Errorf("failed to initialize plugin %s: %w", name, err)
		}

		r.logger.Info("Initialized plugin", zap.String("name", name))
	}

	return nil
}

// StartAll starts all enabled plugins
func (r *PluginRegistry) StartAll() error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Iterate over the map of enabled plugins
	for name, enabled := range r.config.Plugins.Enabled {
		if !enabled {
			continue // Skip disabled plugins
		}

		plugin, err := r.Get(name)
		if err != nil {
			// Log error but continue, maybe plugin wasn't registered
			r.logger.Error("Plugin defined in config but not found in registry", zap.String("name", name), zap.Error(err))
			continue
		}

		if err := plugin.Start(); err != nil {
			return fmt.Errorf("failed to start plugin %s: %w", name, err)
		}

		r.logger.Info("Started plugin", zap.String("name", name))
	}

	return nil
}

// StopAll stops all enabled plugins
func (r *PluginRegistry) StopAll() error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Iterate over the map of enabled plugins
	for name, enabled := range r.config.Plugins.Enabled {
		if !enabled {
			continue // Skip disabled plugins
		}

		plugin, err := r.Get(name)
		if err != nil {
			// Log error but continue, maybe plugin wasn't registered
			r.logger.Error("Plugin defined in config but not found in registry", zap.String("name", name), zap.Error(err))
			continue
		}

		if err := plugin.Stop(); err != nil {
			r.logger.Error("Failed to stop plugin", zap.String("name", name), zap.Error(err))
		} else {
			r.logger.Info("Stopped plugin", zap.String("name", name))
		}
	}

	return nil
}
