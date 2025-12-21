package server

import (
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/axiomod/axiomod/framework/config"
	"github.com/axiomod/axiomod/framework/health"
	"github.com/axiomod/axiomod/framework/middleware"
	"github.com/axiomod/axiomod/platform/observability"
	"go.opentelemetry.io/otel/trace"

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
	metricsMid := middleware.NewMetricsMiddleware(metrics)
	tracingMid := middleware.NewTracingMiddleware(&observability.Tracer{
		Tracer: trace.NewNoopTracerProvider().Tracer("test"),
	})
	h := health.New(logger)

	srv := NewHTTPServer(cfg, logger, metrics, metricsMid, tracingMid, h)

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
				if tt.name == "Health (Legacy)" {
					assert.JSONEq(t, tt.body, string(body))
				} else {
					assert.Contains(t, string(body), `"status":"UP"`)
				}
				resp.Body.Close()
			})
		}
	})
}
