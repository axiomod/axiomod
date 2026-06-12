package server

import (
	"context"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/axiomod/axiomod/framework/config"
	"github.com/axiomod/axiomod/framework/health"
	"github.com/axiomod/axiomod/framework/middleware"
	"github.com/axiomod/axiomod/platform/observability"
	"go.opentelemetry.io/otel/trace"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

func TestRegisterHTTPServerFailsOnOccupiedPort(t *testing.T) {
	// Occupy a port so the server cannot bind to it
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer ln.Close()
	port := ln.Addr().(*net.TCPAddr).Port

	cfg := &config.Config{
		App: config.AppConfig{Name: "test-app"},
		HTTP: config.HTTPConfig{
			Host: "127.0.0.1",
			Port: port,
		},
		Observability: config.ObservabilityConfig{LogLevel: "error"},
	}

	logger, _ := observability.NewLogger(cfg)
	srv := &HTTPServer{App: nil, Config: cfg, Logger: logger}

	app := fxtest.New(t,
		fx.Supply(srv),
		fx.Invoke(RegisterHTTPServer),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = app.Start(ctx)
	assert.Error(t, err, "startup must fail when the HTTP port is unavailable")
	assert.Contains(t, err.Error(), "failed to bind HTTP server")
}

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
