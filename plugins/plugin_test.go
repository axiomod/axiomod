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
	cfg := &config.Config{
		Plugins: config.PluginsConfig{
			Enabled: map[string]bool{
				"mock": true,
			},
			Settings: map[string]map[string]interface{}{
				"mock": {"key": "value"},
			},
		},
	}

	obsCfg := &config.Config{}
	logger, _ := observability.NewLogger(obsCfg)

	t.Run("Register and Lifecycle", func(t *testing.T) {
		metrics, _ := observability.NewMetrics(obsCfg, logger)
		registry, err := NewPluginRegistry(cfg, logger, metrics, nil)
		assert.NoError(t, err)

		mock := &mockPlugin{name: "mock"}
		registry.Register(mock)

		// Check if initialized during NewPluginRegistry (actually initializeEnabledPlugins is called in NewPluginRegistry)
		// But in NewPluginRegistry, registerBuiltInPlugins is called first.
		// If we register after NewPluginRegistry, we need to call initializeEnabledPlugins manually or mock it.

		// Let's re-test with Register called BEFORE initialization logic if possible,
		// but PluginRegistry currently calls registration in constructor.

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

	t.Run("Get Plugin", func(t *testing.T) {
		metrics, _ := observability.NewMetrics(obsCfg, logger)
		registry, _ := NewPluginRegistry(cfg, logger, metrics, nil)
		mock := &mockPlugin{name: "mock-2"}
		registry.Register(mock)

		p, err := registry.Get("mock-2")
		assert.NoError(t, err)
		assert.Equal(t, mock, p)

		_, err = registry.Get("non-existent")
		assert.Error(t, err)
	})
}
