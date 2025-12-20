package observability

import (
	"net/http"

	"axiomod/internal/framework/config"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	// Temporarily comment out problematic imports
	// "go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	// "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	// "go.opentelemetry.io/otel/sdk/resource"
	// sdktrace "go.opentelemetry.io/otel/sdk/trace"
	// semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Module provides the fx options for the observability module
var Module = fx.Options(
	fx.Provide(NewLogger),
	fx.Provide(NewTracer),
	fx.Provide(NewMetrics),
)

// Logger is a wrapper around zap.Logger
type Logger struct {
	*zap.Logger
}

// NewLogger creates a new logger
func NewLogger(cfg *config.Config) (*Logger, error) {
	// Configure logger based on config
	var logLevel zapcore.Level
	if err := logLevel.UnmarshalText([]byte(cfg.Observability.LogLevel)); err != nil {
		logLevel = zapcore.InfoLevel
	}

	var zapConfig zap.Config
	if cfg.Observability.LogFormat == "json" {
		zapConfig = zap.NewProductionConfig()
	} else {
		zapConfig = zap.NewDevelopmentConfig()
	}

	zapConfig.Level = zap.NewAtomicLevelAt(logLevel)

	logger, err := zapConfig.Build(
		zap.AddCallerSkip(1),
		zap.Fields(
			zap.String("service", cfg.App.Name),
			zap.String("environment", cfg.App.Environment),
		),
	)
	if err != nil {
		return nil, err
	}

	return &Logger{Logger: logger}, nil
}

// Tracer is a wrapper around trace.Tracer
type Tracer struct {
	tracer trace.Tracer
}

// NewTracer creates a new tracer
func NewTracer(cfg *config.Config, logger *Logger) (*Tracer, error) {
	// Always return a no-op tracer for now to avoid dependency conflicts
	// and stay within the scope of stabilizing current features
	logger.Info("Using no-op tracer for stabilization phase")
	return &Tracer{tracer: trace.NewNoopTracerProvider().Tracer("")}, nil
}

// Metrics is a wrapper around prometheus.Registry
type Metrics struct {
	Registry *prometheus.Registry
	Handler  http.Handler
}

// NewMetrics creates a new metrics registry
func NewMetrics(cfg *config.Config, logger *Logger) (*Metrics, error) {
	metricsEnabled := cfg.Observability.MetricsEnabled
	metricsPort := cfg.Observability.MetricsPort

	if !metricsEnabled {
		return &Metrics{
			Registry: prometheus.NewRegistry(),
			Handler:  promhttp.Handler(),
		}, nil
	}

	registry := prometheus.NewRegistry()
	registry.MustRegister(prometheus.NewGoCollector())
	registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))

	handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})

	logger.Info("Metrics initialized", zap.Int("port", metricsPort))
	return &Metrics{
		Registry: registry,
		Handler:  handler,
	}, nil
}
