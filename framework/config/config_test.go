package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigLoad(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configContent := `
app:
  name: test-app
  environment: test
  version: 1.0.0
  debug: true
auth:
  oidc:
    issuerURL: https://example.com
    clientID: test-client
    jwksCacheTTL: 60
`
	configPath := filepath.Join(tempDir, "service_default.yaml")
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	assert.NoError(t, err)

	t.Run("Load valid config", func(t *testing.T) {
		cfg, err := Load(configPath)
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, "test-app", cfg.App.Name)
		assert.Equal(t, "test", cfg.App.Environment)
		assert.Equal(t, "1.0.0", cfg.App.Version)
		assert.True(t, cfg.App.Debug)
		assert.Equal(t, "https://example.com", cfg.Auth.OIDC.IssuerURL)
		assert.Equal(t, "test-client", cfg.Auth.OIDC.ClientID)
		assert.Equal(t, 60, cfg.Auth.OIDC.JWKSCacheTTL)
	})

	t.Run("Load non-existent config", func(t *testing.T) {
		_, err := Load("non-existent.yaml")
		assert.Error(t, err)
	})
}

func TestViperProvider(t *testing.T) {
	tempDir := t.TempDir()
	configContent := `
test:
  key: value
  int: 123
  bool: true
`
	configPath := filepath.Join(tempDir, "config.yaml")
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	assert.NoError(t, err)

	provider, err := NewViperProvider(tempDir, "config", "yaml")
	assert.NoError(t, err)

	t.Run("GetString", func(t *testing.T) {
		assert.Equal(t, "value", provider.GetString("test.key"))
	})

	t.Run("GetInt", func(t *testing.T) {
		assert.Equal(t, 123, provider.GetInt("test.int"))
	})

	t.Run("GetBool", func(t *testing.T) {
		assert.True(t, provider.GetBool("test.bool"))
	})

	t.Run("IsSet", func(t *testing.T) {
		assert.True(t, provider.IsSet("test.key"))
		assert.False(t, provider.IsSet("test.missing"))
	})
}

func TestLoadHelpers(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("LoadConfigFile", func(t *testing.T) {
		path := filepath.Join(tempDir, "test_load.yaml")
		_ = os.WriteFile(path, []byte("key: value"), 0644)
		p, err := LoadConfigFile(path)
		assert.NoError(t, err)
		assert.NotNil(t, p)
	})

	t.Run("LoadServiceConfig", func(t *testing.T) {
		path := filepath.Join(tempDir, "service_default.yaml")
		_ = os.WriteFile(path, []byte("app: { name: test }"), 0644)

		// Mock CWD or paths
		origWd, _ := os.Getwd()
		_ = os.Chdir(tempDir)
		defer os.Chdir(origWd)

		p, err := LoadServiceConfig()
		assert.NoError(t, err)
		assert.True(t, p.IsSet("app.name"))
	})

	t.Run("LoadPluginConfig", func(t *testing.T) {
		path := filepath.Join(tempDir, "plugin_settings.yaml")
		_ = os.WriteFile(path, []byte("plugins: { enabled: [] }"), 0644)

		origWd, _ := os.Getwd()
		_ = os.Chdir(tempDir)
		defer os.Chdir(origWd)

		p, err := LoadPluginConfig()
		assert.NoError(t, err)
		assert.True(t, p.IsSet("plugins"))
	})

	t.Run("LoadCLIConfig", func(t *testing.T) {
		path := filepath.Join(tempDir, "cli_config.yaml")
		_ = os.WriteFile(path, []byte("cli: { debug: true }"), 0644)

		origWd, _ := os.Getwd()
		_ = os.Chdir(tempDir)
		defer os.Chdir(origWd)

		p, err := LoadCLIConfig()
		assert.NoError(t, err)
		assert.True(t, p.IsSet("cli.debug"))
	})
}
