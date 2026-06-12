package observability

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/axiomod/axiomod/framework/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap/zapcore"
)

// testConfig returns a quiet baseline config for observability tests.
func testConfig() *config.Config {
	return &config.Config{
		App: config.AppConfig{
			Name:        "test-service",
			Environment: "test",
		},
		Observability: config.ObservabilityConfig{
			LogLevel:  "error",
			LogFormat: "json",
		},
	}
}

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name         string
		logLevel     string
		logFormat    string
		debugEnabled bool
		infoEnabled  bool
	}{
		{"json format debug level", "debug", "json", true, true},
		{"json format error level", "error", "json", false, false},
		{"console format info level", "info", "console", false, true},
		{"unknown format falls back to development", "info", "not-a-format", false, true},
		{"invalid level falls back to info", "not-a-level", "json", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := testConfig()
			cfg.Observability.LogLevel = tt.logLevel
			cfg.Observability.LogFormat = tt.logFormat

			logger, err := NewLogger(cfg)
			require.NoError(t, err)
			require.NotNil(t, logger)
			require.NotNil(t, logger.Logger)

			assert.Equal(t, tt.debugEnabled, logger.Core().Enabled(zapcore.DebugLevel))
			assert.Equal(t, tt.infoEnabled, logger.Core().Enabled(zapcore.InfoLevel))
		})
	}
}

func TestNewMetrics_Disabled(t *testing.T) {
	cfg := testConfig()
	cfg.Observability.MetricsEnabled = false

	logger, err := NewLogger(cfg)
	require.NoError(t, err)

	metrics, err := NewMetrics(cfg, logger)
	require.NoError(t, err)
	require.NotNil(t, metrics)

	assert.NotNil(t, metrics.Registry)
	assert.NotNil(t, metrics.Handler)
	assert.Nil(t, metrics.HTTPRequestsTotal)
	assert.Nil(t, metrics.HTTPRequestDuration)
	assert.Nil(t, metrics.GRPCRequestsTotal)
	assert.Nil(t, metrics.GRPCRequestDuration)
	assert.Nil(t, metrics.DBQueryDuration)
}

func TestNewMetrics_Enabled(t *testing.T) {
	cfg := testConfig()
	cfg.Observability.MetricsEnabled = true
	cfg.Observability.MetricsPort = 9100

	logger, err := NewLogger(cfg)
	require.NoError(t, err)

	metrics, err := NewMetrics(cfg, logger)
	require.NoError(t, err)
	require.NotNil(t, metrics)

	assert.NotNil(t, metrics.Registry)
	assert.NotNil(t, metrics.Handler)
	require.NotNil(t, metrics.HTTPRequestsTotal)
	require.NotNil(t, metrics.HTTPRequestDuration)
	require.NotNil(t, metrics.GRPCRequestsTotal)
	require.NotNil(t, metrics.GRPCRequestDuration)
	require.NotNil(t, metrics.DBQueryDuration)
}

func TestNewMetrics_VectorLabelCounts(t *testing.T) {
	cfg := testConfig()
	cfg.Observability.MetricsEnabled = true

	logger, err := NewLogger(cfg)
	require.NoError(t, err)
	metrics, err := NewMetrics(cfg, logger)
	require.NoError(t, err)

	t.Run("http vectors take method, path, status", func(t *testing.T) {
		_, err := metrics.HTTPRequestsTotal.GetMetricWithLabelValues("GET", "/api", "200")
		assert.NoError(t, err)
		_, err = metrics.HTTPRequestsTotal.GetMetricWithLabelValues("GET", "/api")
		assert.Error(t, err, "wrong label cardinality must be rejected")

		_, err = metrics.HTTPRequestDuration.GetMetricWithLabelValues("GET", "/api", "200")
		assert.NoError(t, err)
	})

	t.Run("grpc vectors take service, method, status", func(t *testing.T) {
		_, err := metrics.GRPCRequestsTotal.GetMetricWithLabelValues("pkg.Service", "Method", "OK")
		assert.NoError(t, err)
		_, err = metrics.GRPCRequestDuration.GetMetricWithLabelValues("pkg.Service", "Method", "OK")
		assert.NoError(t, err)
		_, err = metrics.GRPCRequestsTotal.GetMetricWithLabelValues("pkg.Service")
		assert.Error(t, err, "wrong label cardinality must be rejected")
	})

	t.Run("db vector takes query_type, status", func(t *testing.T) {
		_, err := metrics.DBQueryDuration.GetMetricWithLabelValues("select", "success")
		assert.NoError(t, err)
		_, err = metrics.DBQueryDuration.GetMetricWithLabelValues("select", "success", "extra")
		assert.Error(t, err, "wrong label cardinality must be rejected")
	})
}

func TestNewMetrics_RegistryGathersVectors(t *testing.T) {
	cfg := testConfig()
	cfg.Observability.MetricsEnabled = true

	logger, err := NewLogger(cfg)
	require.NoError(t, err)
	metrics, err := NewMetrics(cfg, logger)
	require.NoError(t, err)

	metrics.HTTPRequestsTotal.WithLabelValues("GET", "/api", "200").Inc()
	metrics.HTTPRequestDuration.WithLabelValues("GET", "/api", "200").Observe(0.05)
	metrics.GRPCRequestsTotal.WithLabelValues("pkg.Service", "Method", "OK").Inc()
	metrics.GRPCRequestDuration.WithLabelValues("pkg.Service", "Method", "OK").Observe(0.05)
	metrics.DBQueryDuration.WithLabelValues("select", "success").Observe(0.01)

	families, err := metrics.Registry.Gather()
	require.NoError(t, err)

	names := make(map[string]bool, len(families))
	for _, family := range families {
		names[family.GetName()] = true
	}

	for _, expected := range []string{
		"http_requests_total",
		"http_request_duration_seconds",
		"grpc_requests_total",
		"grpc_request_duration_seconds",
		"db_query_duration_seconds",
	} {
		assert.True(t, names[expected], "expected metric family %q to be registered", expected)
	}
}

func TestMetricsHandler_ServesPrometheusFormat(t *testing.T) {
	cfg := testConfig()
	cfg.Observability.MetricsEnabled = true

	logger, err := NewLogger(cfg)
	require.NoError(t, err)
	metrics, err := NewMetrics(cfg, logger)
	require.NoError(t, err)

	metrics.HTTPRequestsTotal.WithLabelValues("GET", "/api", "200").Inc()

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	metrics.Handler.ServeHTTP(rec, req)

	resp := rec.Result()
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, string(body), "http_requests_total")
}

func TestNewTracer_Disabled(t *testing.T) {
	cfg := testConfig()
	cfg.Observability.TracingEnabled = false

	logger, err := NewLogger(cfg)
	require.NoError(t, err)

	tracer, err := NewTracer(cfg, logger)
	require.NoError(t, err)
	require.NotNil(t, tracer)

	assert.Nil(t, tracer.Provider, "disabled tracing must not create a provider")
	require.NotNil(t, tracer.Tracer)

	// The no-op tracer must produce non-recording spans without errors.
	_, span := tracer.Tracer.Start(context.Background(), "test-span")
	assert.False(t, span.IsRecording())
	span.End()
}

func TestNewTracer_StdoutExporter(t *testing.T) {
	cfg := testConfig()
	cfg.Observability.TracingEnabled = true
	cfg.Observability.TracingExporterType = "stdout"
	// NeverSample keeps test output free of exported spans.
	cfg.Observability.TracingSamplerRatio = 0

	logger, err := NewLogger(cfg)
	require.NoError(t, err)

	tracer, err := NewTracer(cfg, logger)
	require.NoError(t, err)
	require.NotNil(t, tracer)
	require.NotNil(t, tracer.Provider)
	require.NotNil(t, tracer.Tracer)

	_, span := tracer.Tracer.Start(context.Background(), "test-span")
	span.End()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	require.NoError(t, tracer.Provider.Shutdown(ctx))
}

func TestNewTracer_UnknownExporterFallsBackToStdout(t *testing.T) {
	cfg := testConfig()
	cfg.Observability.TracingEnabled = true
	cfg.Observability.TracingExporterType = "unknown-exporter"
	cfg.Observability.TracingSamplerRatio = 0.5 // Exercise the ratio-based sampler path.

	logger, err := NewLogger(cfg)
	require.NoError(t, err)

	tracer, err := NewTracer(cfg, logger)
	require.NoError(t, err)
	require.NotNil(t, tracer)
	require.NotNil(t, tracer.Provider)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	require.NoError(t, tracer.Provider.Shutdown(ctx))
}

func TestRegisterTracer(t *testing.T) {
	cfg := testConfig()
	logger, err := NewLogger(cfg)
	require.NoError(t, err)

	t.Run("nil provider registers no hooks", func(t *testing.T) {
		lc := fxtest.NewLifecycle(t)
		RegisterTracer(lc, &Tracer{Provider: nil}, logger)
		lc.RequireStart().RequireStop()
	})

	t.Run("real provider is shut down on stop", func(t *testing.T) {
		tracingCfg := testConfig()
		tracingCfg.Observability.TracingEnabled = true
		tracingCfg.Observability.TracingExporterType = "stdout"
		tracingCfg.Observability.TracingSamplerRatio = 0

		tracer, err := NewTracer(tracingCfg, logger)
		require.NoError(t, err)
		require.NotNil(t, tracer.Provider)

		lc := fxtest.NewLifecycle(t)
		RegisterTracer(lc, tracer, logger)
		lc.RequireStart().RequireStop()
	})
}

func TestModule_FxIntegration(t *testing.T) {
	var logger *Logger
	var tracer *Tracer
	var metrics *Metrics

	app := fxtest.New(t,
		fx.Provide(func() *config.Config {
			cfg := testConfig()
			cfg.Observability.MetricsEnabled = true
			return cfg
		}),
		Module,
		fx.Invoke(func(l *Logger, tr *Tracer, m *Metrics) {
			logger = l
			tracer = tr
			metrics = m
		}),
	)
	app.RequireStart()
	defer app.RequireStop()

	assert.NotNil(t, logger)
	assert.NotNil(t, tracer)
	assert.NotNil(t, metrics)
}
