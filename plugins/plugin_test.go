package plugins

import (
	"testing"

	"github.com/axiomod/axiomod/framework/config"
	"github.com/axiomod/axiomod/framework/health"
	"github.com/axiomod/axiomod/platform/observability"

	"github.com/stretchr/testify/assert"
)

type mockPlugin struct {
	name        string
	initialized bool
	started     bool
	stopped     bool
}

func (m *mockPlugin) Name() string { return m.name }
func (m *mockPlugin) Initialize(settings map[string]interface{}, logger *observability.Logger, metrics *observability.Metrics, cfg *config.Config, health *health.Health) error {
	m.initialized = true
	return nil
}
func (m *mockPlugin) Start() error {
	m.started = true
	return nil
}
func (m *mockPlugin) Stop() error {
	m.stopped = true
	return nil
}

func TestPluginRegistry(t *testing.T) {
	obsCfg := &config.Config{}
	logger, _ := observability.NewLogger(obsCfg)

	t.Run("Register and Lifecycle", func(t *testing.T) {
		// Start with no plugins enabled so the constructor succeeds, then
		// register the mock and enable it before running the lifecycle.
		cfg := &config.Config{
			Plugins: config.PluginsConfig{
				Enabled: map[string]bool{},
				Settings: map[string]map[string]interface{}{
					"mock": {"key": "value"},
				},
			},
		}

		metrics, _ := observability.NewMetrics(obsCfg, logger)
		registry, err := NewPluginRegistry(cfg, logger, metrics, nil)
		assert.NoError(t, err)

		mock := &mockPlugin{name: "mock"}
		registry.Register(mock)
		cfg.Plugins.Enabled["mock"] = true

		err = registry.initializeEnabledPlugins()
		assert.NoError(t, err)
		assert.True(t, mock.initialized)

		err = registry.StartAll()
		assert.NoError(t, err)
		assert.True(t, mock.started)

		err = registry.StopAll()
		assert.NoError(t, err)
		assert.True(t, mock.stopped)
	})

	t.Run("Enabled but Unregistered Plugin Fails Construction", func(t *testing.T) {
		cfg := &config.Config{
			Plugins: config.PluginsConfig{
				Enabled: map[string]bool{"no-such-plugin": true},
			},
		}

		metrics, _ := observability.NewMetrics(obsCfg, logger)
		registry, err := NewPluginRegistry(cfg, logger, metrics, nil)
		assert.Error(t, err)
		assert.Nil(t, registry)
		assert.Contains(t, err.Error(), "no-such-plugin")
	})

	t.Run("Get Plugin", func(t *testing.T) {
		cfg := &config.Config{
			Plugins: config.PluginsConfig{Enabled: map[string]bool{}},
		}

		metrics, _ := observability.NewMetrics(obsCfg, logger)
		registry, err := NewPluginRegistry(cfg, logger, metrics, nil)
		assert.NoError(t, err)

		mock := &mockPlugin{name: "mock-2"}
		registry.Register(mock)

		p, err := registry.Get("mock-2")
		assert.NoError(t, err)
		assert.Equal(t, mock, p)

		_, err = registry.Get("non-existent")
		assert.Error(t, err)
	})
}
