package server

import (
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/axiomod/axiomod/framework/config"
	"github.com/axiomod/axiomod/platform/observability"

	"github.com/stretchr/testify/assert"
)

func TestHTTPServer(t *testing.T) {
	cfg := &config.Config{
		App: config.AppConfig{Name: "test-app"},
		HTTP: config.HTTPConfig{
			Host:         "localhost",
			Port:         8081, // Use different port
			ReadTimeout:  5,
			WriteTimeout: 5,
		},
		Observability: config.ObservabilityConfig{
			LogLevel: "debug",
		},
	}

	logger, _ := observability.NewLogger(cfg)
	metrics, _ := observability.NewMetrics(cfg, logger)

	srv := NewHTTPServer(cfg, logger, metrics)

	t.Run("Health Endpoints", func(t *testing.T) {
		// Run server in background for testing probes
		go func() {
			_ = srv.App.Listen(":8081")
		}()

		// Give server time to start
		time.Sleep(100 * time.Millisecond)
		defer srv.App.Shutdown()

		tests := []struct {
			name   string
			path   string
			status int
			body   string
		}{
			{"Liveness", "/live", http.StatusOK, `{"status":"alive"}`},
			{"Readiness", "/ready", http.StatusOK, `{"status":"ready"}`},
			{"Health (Legacy)", "/health", http.StatusOK, `{"status":"ok"}`},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				resp, err := http.Get("http://localhost:8081" + tt.path)
				assert.NoError(t, err)
				assert.Equal(t, tt.status, resp.StatusCode)

				body, _ := io.ReadAll(resp.Body)
				assert.JSONEq(t, tt.body, string(body))
				resp.Body.Close()
			})
		}
	})
}
