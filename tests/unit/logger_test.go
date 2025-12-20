package unit_test

import (
	"testing"

	"axiomod/internal/framework/config"
	"axiomod/internal/platform/observability"

	"github.com/stretchr/testify/assert"
)

func TestLoggerInitialization(t *testing.T) {
	// Create a minimal config for testing
	cfg := &config.Config{
		App: config.AppConfig{
			Name:        "test-app",
			Environment: "test",
			Version:     "1.0",
		},
		Observability: config.ObservabilityConfig{
			LogLevel:  "info",
			LogFormat: "text",
		},
	}

	// Test logger initialization using the observability package
	log, err := observability.NewLogger(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, log)
	log.Info("Logger initialized successfully for testing")
}
