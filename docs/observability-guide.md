# Observability Guide

## Introduction

Observability is a critical aspect of modern applications, especially in distributed systems. The Enterprise Axiomod provides built-in support for observability through logging, metrics, and tracing. This guide will help you understand how to use these features effectively.

## Logging

The framework uses Uber's Zap for structured logging, which provides high-performance, structured logging with minimal allocations.

### Configuration

Logging can be configured in the `config.yaml` file:

```yaml
observability:
  logLevel: info  # debug, info, warn, error, dpanic, panic, fatal
  logFormat: json # json or console
```

Or through environment variables:

```bash
export LOG_LEVEL=info
export LOG_FORMAT=json
```

### Usage

The logger is injected into components through dependency injection:

```go
type MyService struct {
    logger *observability.Logger
}

func NewMyService(logger *observability.Logger) *MyService {
    return &MyService{
        logger: logger,
    }
}

func (s *MyService) DoSomething() {
    s.logger.Info("Doing something",
        zap.String("key", "value"),
        zap.Int("count", 42),
    )
    
    // With error
    if err := someOperation(); err != nil {
        s.logger.Error("Operation failed", zap.Error(err))
    }
}
```

### Log Levels

The framework supports the following log levels:

- **Debug**: Detailed information, typically of interest only when diagnosing problems
- **Info**: Confirmation that things are working as expected
- **Warn**: Indication that something unexpected happened, but the application can continue
- **Error**: Due to a more serious problem, the application has not been able to perform a function
- **DPanic**: Critical errors that should typically trigger a panic in development
- **Panic**: Critical errors that trigger a panic
- **Fatal**: Critical errors that cause the application to exit

### Structured Logging

Always use structured logging with fields to provide context:

```go
// Good
logger.Info("User created", 
    zap.String("user_id", user.ID),
    zap.String("email", user.Email),
)

// Bad
logger.Info(fmt.Sprintf("User created with ID %s and email %s", user.ID, user.Email))
```

## Metrics

The framework uses Prometheus for metrics collection, which provides a powerful monitoring system and time series database.

### Configuration

Metrics can be configured in the `config.yaml` file:

```yaml
observability:
  metricsEnabled: true
  metricsPort: 9100
```

Or through environment variables:

```bash
export METRICS_ENABLED=true
export METRICS_PORT=9100
```

### Usage

The metrics registry is injected into components through dependency injection:

```go
type MyService struct {
    metrics *observability.Metrics
    requestCounter prometheus.Counter
}

func NewMyService(metrics *observability.Metrics) *MyService {
    // Create a counter
    requestCounter := prometheus.NewCounter(prometheus.CounterOpts{
        Name: "my_service_requests_total",
        Help: "Total number of requests processed by MyService",
    })
    
    // Register the counter with the registry
    metrics.Registry.MustRegister(requestCounter)
    
    return &MyService{
        metrics: metrics,
        requestCounter: requestCounter,
    }
}

func (s *MyService) HandleRequest() {
    // Increment the counter
    s.requestCounter.Inc()
    
    // Process the request...
}
```

### Common Metric Types

The framework supports the following metric types:

- **Counter**: A cumulative metric that represents a single monotonically increasing counter
- **Gauge**: A metric that represents a single numerical value that can arbitrarily go up and down
- **Histogram**: Samples observations and counts them in configurable buckets
- **Summary**: Similar to a histogram, but also calculates quantiles over a sliding time window

### Metric Naming Conventions

Follow these conventions for metric names:

- Use snake_case for metric names
- Use a prefix for your service or module
- Be descriptive but concise
- Include units in the name if applicable

Examples:
- `http_requests_total`
- `database_connections_active`
- `request_duration_seconds`

## Tracing

The framework uses OpenTelemetry for distributed tracing, which provides a vendor-neutral API for tracing.

### Configuration

Tracing can be configured in the `config.yaml` file:

```yaml
observability:
  tracingEnabled: true
  tracingServiceName: axiomod
  tracingExporterType: jaeger
  tracingExporterURL: http://jaeger:14268/api/traces
```

Or through environment variables:

```bash
export TRACING_ENABLED=true
export TRACING_SERVICE_NAME=axiomod
export TRACING_EXPORTER_TYPE=jaeger
export TRACING_EXPORTER_URL=http://jaeger:14268/api/traces
```

### Usage

The tracer is injected into components through dependency injection:

```go
type MyService struct {
    tracer *observability.Tracer
}

func NewMyService(tracer *observability.Tracer) *MyService {
    return &MyService{
        tracer: tracer,
    }
}

func (s *MyService) DoSomething(ctx context.Context) {
    // Create a span
    ctx, span := s.tracer.Start(ctx, "MyService.DoSomething")
    defer span.End()
    
    // Add attributes to the span
    span.SetAttributes(attribute.String("key", "value"))
    
    // Record an event
    span.AddEvent("Processing started")
    
    // Process...
    
    // Record another event
    span.AddEvent("Processing completed")
}
```

### Propagating Context

When making calls to other services, propagate the context:

```go
func (s *MyService) CallOtherService(ctx context.Context) {
    // Create a span
    ctx, span := s.tracer.Start(ctx, "MyService.CallOtherService")
    defer span.End()
    
    // Make the call with the context
    response, err := s.otherService.DoSomething(ctx, request)
    
    if err != nil {
        // Record the error
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return
    }
    
    // Process the response...
}
```

### Sampling

In production environments, you may want to sample traces to reduce the volume of data:

```yaml
observability:
  tracingSamplingRatio: 0.1  # Sample 10% of traces
```

## Health Checks

The framework provides health check endpoints to monitor the health of the application.

### Configuration

Health checks are enabled by default and can be accessed at the `/health` endpoint.

### Types of Health Checks

The framework provides two types of health checks:

- **Liveness**: Indicates if the application is running
- **Readiness**: Indicates if the application is ready to accept requests

### Custom Health Checks

You can add custom health checks:

```go
type MyService struct {
    health *health.Health
}

func NewMyService(health *health.Health) *MyService {
    // Register a health check
    health.RegisterCheck("database", func() error {
        // Check database connection
        if !isDatabaseConnected() {
            return errors.New("database not connected")
        }
        return nil
    })
    
    return &MyService{
        health: health,
    }
}
```

## Integrating with Monitoring Systems

### Prometheus and Grafana

1. Configure Prometheus to scrape metrics from your application:

```yaml
scrape_configs:
  - job_name: 'axiomod'
    scrape_interval: 15s
    static_configs:
      - targets: ['axiomod:9100']
```

2. Create Grafana dashboards to visualize the metrics.

### ELK Stack

1. Configure Logstash to collect logs:

```
input {
  tcp {
    port => 5000
    codec => json
  }
}

filter {
  json {
    source => "message"
  }
}

output {
  elasticsearch {
    hosts => ["elasticsearch:9200"]
    index => "axiomod-%{+YYYY.MM.dd}"
  }
}
```

2. Create Kibana visualizations to analyze the logs.

### Jaeger

1. Configure the application to send traces to Jaeger:

```yaml
observability:
  tracingExporterType: jaeger
  tracingExporterURL: http://jaeger:14268/api/traces
```

2. Use the Jaeger UI to view and analyze traces.

## Best Practices

### 1. Use Structured Logging

Always use structured logging with fields to provide context.

### 2. Use Meaningful Metric Names

Choose metric names that are descriptive and follow naming conventions.

### 3. Add Context to Spans

Add attributes and events to spans to provide context for troubleshooting.

### 4. Propagate Context

Always propagate the context when making calls to other services.

### 5. Monitor Health Checks

Set up alerts based on health check endpoints to detect issues early.

### 6. Use Sampling in Production

Use sampling to reduce the volume of trace data in production environments.

### 7. Correlate Logs, Metrics, and Traces

Use correlation IDs to correlate logs, metrics, and traces for a complete view of the system.

## Conclusion

Observability is a critical aspect of modern applications. By using the logging, metrics, and tracing features provided by the Enterprise Axiomod, you can gain insights into the behavior of your application and quickly identify and resolve issues.
