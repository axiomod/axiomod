package plugins

import (
	"context"
	"fmt"
	"sync"

	"github.com/axiomod/axiomod/framework/config"
	"github.com/axiomod/axiomod/framework/health"
	"github.com/axiomod/axiomod/platform/observability"

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

	// Initialize initializes the plugin with the given configuration, logger, and metrics
	Initialize(config map[string]interface{}, logger *observability.Logger, metrics *observability.Metrics, cfg *config.Config, health *health.Health) error

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
	metrics *observability.Metrics
	health  *health.Health
	mu      sync.RWMutex
}

// NewPluginRegistry creates a new plugin registry
func NewPluginRegistry(cfg *config.Config, logger *observability.Logger, metrics *observability.Metrics, health *health.Health) (*PluginRegistry, error) {
	registry := &PluginRegistry{
		plugins: make(map[string]Plugin),
		config:  cfg,
		logger:  logger,
		metrics: metrics,
		health:  health,
	}

	// Register built-in plugins. Initialization is deferred to StartAll so
	// that plugins registered after construction (e.g. via fx.Invoke) are
	// initialized too.
	registry.registerBuiltInPlugins()

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

	return r.get(name)
}

// get looks up a plugin without locking; callers must hold r.mu.
func (r *PluginRegistry) get(name string) (Plugin, error) {
	plugin, ok := r.plugins[name]
	if !ok {
		return nil, fmt.Errorf("plugin not found: %s", name)
	}

	return plugin, nil
}

// StartAll initializes and starts all enabled plugins. It runs after all
// registrations (built-in and fx-invoked) so every enabled plugin is
// initialized exactly once before it is started.
func (r *PluginRegistry) StartAll() error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Iterate over the map of enabled plugins
	for name, enabled := range r.config.Plugins.Enabled {
		if !enabled {
			continue // Skip disabled plugins
		}

		plugin, err := r.get(name)
		if err != nil {
			// Fail fast: an enabled plugin that is not registered is a
			// configuration error (e.g. a typo in plugins.enabled).
			return fmt.Errorf("plugin %q is enabled in config but not registered: %w", name, err)
		}

		// Get plugin settings
		pluginSettings, ok := r.config.Plugins.Settings[name]
		if !ok {
			pluginSettings = make(map[string]interface{}) // Use empty settings if none found
		}

		if err := plugin.Initialize(pluginSettings, r.logger, r.metrics, r.config, r.health); err != nil {
			return fmt.Errorf("failed to initialize plugin %s: %w", name, err)
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

		plugin, err := r.get(name)
		if err != nil {
			// During shutdown, log and continue so remaining plugins still stop.
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
