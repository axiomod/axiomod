package observability

import (
	"context"
	"fmt"
	"net/http"

	"github.com/axiomod/axiomod/framework/config"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
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
	fx.Invoke(RegisterTracer),
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
	Tracer   trace.Tracer
	Provider *sdktrace.TracerProvider
}

// NewTracer creates a new tracer
func NewTracer(cfg *config.Config, logger *Logger) (*Tracer, error) {
	if !cfg.Observability.TracingEnabled {
		logger.Info("Tracing is disabled, using no-op tracer")
		return &Tracer{
			Tracer:   trace.NewNoopTracerProvider().Tracer(cfg.App.Name),
			Provider: nil,
		}, nil
	}

	tp, err := initTracer(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize tracer: %w", err)
	}

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	logger.Info("Tracer initialized", zap.String("exporter", cfg.Observability.TracingExporterType))

	return &Tracer{
		Tracer:   tp.Tracer(cfg.App.Name),
		Provider: tp,
	}, nil
}

// initTracer initializes the OpenTelemetry tracer provider
func initTracer(cfg *config.Config) (*sdktrace.TracerProvider, error) {
	var exporter sdktrace.SpanExporter
	var err error

	ctx := context.Background()

	switch cfg.Observability.TracingExporterType {
	case "jaeger":
		exporter, err = jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(cfg.Observability.TracingURL)))
	case "otlp":
		// Assume OTLP over GRPC for now, can be made configurable
		exporter, err = otlptracegrpc.New(ctx, otlptracegrpc.WithEndpoint(cfg.Observability.TracingURL), otlptracegrpc.WithInsecure())
	case "stdout":
		exporter, err = stdouttrace.New(stdouttrace.WithPrettyPrint())
	default:
		exporter, err = stdouttrace.New()
	}

	if err != nil {
		return nil, err
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(cfg.App.Name),
			semconv.DeploymentEnvironmentKey.String(cfg.App.Environment),
		),
	)
	if err != nil {
		return nil, err
	}

	sampler := sdktrace.AlwaysSample()
	if cfg.Observability.TracingSamplerRatio > 0 && cfg.Observability.TracingSamplerRatio < 1 {
		sampler = sdktrace.ParentBased(sdktrace.TraceIDRatioBased(cfg.Observability.TracingSamplerRatio))
	} else if cfg.Observability.TracingSamplerRatio == 0 {
		sampler = sdktrace.NeverSample()
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sampler),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	return tp, nil
}

// RegisterTracer registers the tracer with the fx lifecycle
func RegisterTracer(lc fx.Lifecycle, tracer *Tracer, logger *Logger) {
	if tracer.Provider == nil {
		return
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			logger.Info("Shutting down tracer provider")
			return tracer.Provider.Shutdown(ctx)
		},
	})
}

// Metrics is a wrapper around prometheus.Registry
type Metrics struct {
	Registry            *prometheus.Registry
	Handler             http.Handler
	HTTPRequestsTotal   *prometheus.CounterVec
	HTTPRequestDuration *prometheus.HistogramVec
	GRPCRequestsTotal   *prometheus.CounterVec
	GRPCRequestDuration *prometheus.HistogramVec
	DBQueryDuration     *prometheus.HistogramVec
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

	httpRequestsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)
	httpRequestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)

	grpcRequestsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_requests_total",
			Help: "Total number of gRPC requests",
		},
		[]string{"service", "method", "status"},
	)
	grpcRequestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "grpc_request_duration_seconds",
			Help:    "Duration of gRPC requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "method", "status"},
	)

	dbQueryDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "Duration of database queries in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"query_type", "status"},
	)

	registry.MustRegister(httpRequestsTotal)
	registry.MustRegister(httpRequestDuration)
	registry.MustRegister(grpcRequestsTotal)
	registry.MustRegister(grpcRequestDuration)
	registry.MustRegister(dbQueryDuration)

	handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})

	logger.Info("Metrics initialized", zap.Int("port", metricsPort))
	return &Metrics{
		Registry:            registry,
		Handler:             handler,
		HTTPRequestsTotal:   httpRequestsTotal,
		HTTPRequestDuration: httpRequestDuration,
		GRPCRequestsTotal:   grpcRequestsTotal,
		GRPCRequestDuration: grpcRequestDuration,
		DBQueryDuration:     dbQueryDuration,
	}, nil
}
