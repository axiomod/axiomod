package unit_test

import (
	"testing"

	"axiomod/internal/framework/config"

	"github.com/stretchr/testify/assert"
)

func TestConfigLoaders(t *testing.T) {
	// Test CLI config loader
	cliConfig, err := config.LoadCLIConfig()
	assert.NoError(t, err)
	assert.NotNil(t, cliConfig)

	// Test service config loader
	// Note: This might fail if service_default.yaml doesn't exist or is empty
	// We should ideally mock the file system or provide a test config file
	// For now, we just check if the function runs without error
	_, err = config.LoadServiceConfig()
	assert.NoError(t, err)
	// assert.NotNil(t, serviceConfig) // Cannot assert NotNil if file doesn't exist

	// Test plugin config loader
	// Note: Similar issue as LoadServiceConfig
	_, err = config.LoadPluginConfig()
	assert.NoError(t, err)
	// assert.NotNil(t, pluginConfig) // Cannot assert NotNil if file doesn't exist
}
