// Package example_plugin is the reference implementation of the Axiomod
// Plugin interface (plugins/plugin.go). Copy it as a starting point for new
// plugins; docs/plugin-development-guide.md walks through this exact code.
package example_plugin

import (
	"fmt"

	"github.com/axiomod/axiomod/framework/config"
	"github.com/axiomod/axiomod/framework/health"
	"github.com/axiomod/axiomod/framework/observability"

	"go.uber.org/zap"
)

// ExamplePlugin demonstrates a complete plugin: configuration handling,
// health-check registration, and lifecycle management.
type ExamplePlugin struct {
	settings map[string]interface{}
	logger   *observability.Logger
	metrics  *observability.Metrics
	greeting string
	active   bool
}

// Name returns the unique, lowercase plugin name used as the key under
// plugins.enabled and plugins.settings in the service configuration.
func (p *ExamplePlugin) Name() string {
	return "example"
}

// Initialize receives the plugin's settings block (plugins.settings.example),
// the shared observability components, the full application config, and the
// health registry. Register health checks here; defer external connections
// to Start.
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

	// Read plugin settings with sane defaults.
	p.greeting = "Hello"
	if g, ok := settings["greeting"].(string); ok && g != "" {
		p.greeting = g
	}

	// Report this plugin's state through /ready and /health.
	h.RegisterCheck(p.Name(), func() error {
		if !p.active {
			return fmt.Errorf("example plugin is not running")
		}
		return nil
	})

	return nil
}

// Start is called after every enabled plugin has been initialized. Open
// connections and launch background workers here; fail fast on errors.
func (p *ExamplePlugin) Start() error {
	p.active = true
	p.logger.Info("Example plugin started", zap.String("greeting", p.greeting))
	return nil
}

// Stop is called on shutdown (fx OnStop). Release every resource acquired
// in Start.
func (p *ExamplePlugin) Stop() error {
	p.active = false
	p.logger.Info("Example plugin stopped")
	return nil
}

// IsActive reports whether the plugin has been started.
func (p *ExamplePlugin) IsActive() bool {
	return p.active
}
