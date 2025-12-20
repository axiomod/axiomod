package tracing

import (
	"context"
	"fmt"
	"time"

	"axiomod/internal/platform/observability"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// ExporterType represents the type of exporter
type ExporterType string

const (
	// ExporterTypeJaeger is the Jaeger exporter
	ExporterTypeJaeger ExporterType = "jaeger"
	// ExporterTypeOTLP is the OTLP exporter
	ExporterTypeOTLP ExporterType = "otlp"
	// ExporterTypeStdout is the stdout exporter
	ExporterTypeStdout ExporterType = "stdout"
)

// Config contains configuration for the tracer
type Config struct {
	// ServiceName is the name of the service
	ServiceName string
	// ServiceVersion is the version of the service
	ServiceVersion string
	// Environment is the environment (e.g., production, staging, development)
	Environment string
	// ExporterType is the type of exporter to use
	ExporterType ExporterType
	// ExporterEndpoint is the endpoint for the exporter
	ExporterEndpoint string
	// SamplingRatio is the sampling ratio (0.0 to 1.0)
	SamplingRatio float64
	// Attributes are additional attributes to add to the resource
	Attributes map[string]string
}

// DefaultConfig returns the default tracer configuration
func DefaultConfig() *Config {
	return &Config{
		ServiceName:      "axiomod",
		ServiceVersion:   "1.0.0",
		Environment:      "development",
		ExporterType:     ExporterTypeJaeger,
		ExporterEndpoint: "http://localhost:14268/api/traces",
		SamplingRatio:    1.0,
		Attributes:       make(map[string]string),
	}
}

// Tracer provides tracing functionality
type Tracer struct {
	tracer trace.Tracer
	logger *observability.Logger
	config *Config
}

// New creates a new tracer
func New(logger *observability.Logger, config *Config) (*Tracer, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// Create resource
	res, err := createResource(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create exporter
	exporter, err := createExporter(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create exporter: %w", err)
	}

	// Create sampler
	sampler := sdktrace.ParentBased(
		sdktrace.TraceIDRatioBased(config.SamplingRatio),
	)

	// Create trace provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sampler),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	// Set global trace provider
	otel.SetTracerProvider(tp)

	// Set global propagator
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Create tracer
	tracer := tp.Tracer(config.ServiceName)

	logger.Info("Created tracer",
		zap.String("service", config.ServiceName),
		zap.String("exporter", string(config.ExporterType)),
		zap.Float64("sampling_ratio", config.SamplingRatio),
	)

	return &Tracer{
		tracer: tracer,
		logger: logger,
		config: config,
	}, nil
}

// Start starts a new span
func (t *Tracer) Start(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return t.tracer.Start(ctx, name, opts...)
}

// StartWithAttributes starts a new span with attributes
func (t *Tracer) StartWithAttributes(ctx context.Context, name string, attributes map[string]string) (context.Context, trace.Span) {
	// Convert attributes to trace attributes
	attrs := make([]attribute.KeyValue, 0, len(attributes))
	for k, v := range attributes {
		attrs = append(attrs, attribute.String(k, v))
	}

	return t.tracer.Start(ctx, name, trace.WithAttributes(attrs...))
}

// AddEvent adds an event to the current span
func (t *Tracer) AddEvent(ctx context.Context, name string, attributes map[string]string) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	// Convert attributes to trace attributes
	attrs := make([]attribute.KeyValue, 0, len(attributes))
	for k, v := range attributes {
		attrs = append(attrs, attribute.String(k, v))
	}

	span.AddEvent(name, trace.WithAttributes(attrs...))
}

// SetAttributes sets attributes on the current span
func (t *Tracer) SetAttributes(ctx context.Context, attributes map[string]string) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	// Convert attributes to trace attributes
	for k, v := range attributes {
		span.SetAttributes(attribute.String(k, v))
	}
}

// RecordError records an error on the current span
func (t *Tracer) RecordError(ctx context.Context, err error) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	span.RecordError(err)
}

// End ends the current span
func (t *Tracer) End(ctx context.Context) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	span.End()
}

// createResource creates a resource with service information
func createResource(config *Config) (*resource.Resource, error) {
	// Create base resource
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(config.ServiceName),
			semconv.ServiceVersionKey.String(config.ServiceVersion),
			attribute.String("environment", config.Environment),
		),
	)
	if err != nil {
		return nil, err
	}

	// Add custom attributes
	if len(config.Attributes) > 0 {
		attrs := make([]attribute.KeyValue, 0, len(config.Attributes))
		for k, v := range config.Attributes {
			attrs = append(attrs, attribute.String(k, v))
		}
		res, err = resource.Merge(res, resource.NewWithAttributes(semconv.SchemaURL, attrs...))
		if err != nil {
			return nil, err
		}
	}

	return res, nil
}

// createExporter creates an exporter based on the configuration
func createExporter(config *Config) (sdktrace.SpanExporter, error) {
	// ctx := context.Background() // Commented out unused variable

	// Temporarily use stdout exporter to avoid OTLP dependency issues
	fmt.Println("Warning: Using stdout tracer due to OTLP dependency issues. Configure Jaeger or fix OTLP dependencies for production.")
	return stdouttrace.New(
		stdouttrace.WithPrettyPrint(),
		stdouttrace.WithoutTimestamps(),
	)

	/* Original implementation with OTLP - commented out due to dependency conflicts
	 switch config.ExporterType {
	 case ExporterTypeJaeger:
		 return jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(config.ExporterEndpoint)))
	 case ExporterTypeOTLP:
		 // Check if endpoint is HTTP or gRPC
		 if isHTTPEndpoint(config.ExporterEndpoint) {
			 return otlptrace.New(
				 ctx,
				 otlptracehttp.NewClient(
					 otlptracehttp.WithEndpoint(config.ExporterEndpoint),
					 otlptracehttp.WithInsecure(),
				 ),
			 )
		 }
		 return otlptrace.New(
			 ctx,
			 otlptracegrpc.NewClient(
				 otlptracegrpc.WithEndpoint(config.ExporterEndpoint),
				 otlptracegrpc.WithInsecure(),
			 ),
		 )
	 case ExporterTypeStdout:
		 return stdouttrace.New(
			 stdouttrace.WithPrettyPrint(),
			 stdouttrace.WithoutTimestamps(),
		 )
	 default:
		 return nil, fmt.Errorf("unsupported exporter type: %s", config.ExporterType)
	 }
	*/
}

// isHTTPEndpoint checks if an endpoint is HTTP
func isHTTPEndpoint(endpoint string) bool {
	return len(endpoint) >= 4 && (endpoint[:4] == "http" || endpoint[:4] == "HTTP")
}

// WithSpan wraps a function with a span
func WithSpan(ctx context.Context, tracer *Tracer, name string, fn func(context.Context) error) error {
	ctx, span := tracer.Start(ctx, name)
	defer span.End()

	return fn(ctx)
}

// WithSpanAndTimeout wraps a function with a span and timeout
func WithSpanAndTimeout(ctx context.Context, tracer *Tracer, name string, timeout time.Duration, fn func(context.Context) error) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ctx, span := tracer.Start(ctx, name)
	defer span.End()

	return fn(ctx)
}
