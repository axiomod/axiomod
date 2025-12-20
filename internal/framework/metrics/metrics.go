package metrics

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"axiomod/internal/platform/observability"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// Registry is a wrapper around prometheus.Registry
type Registry struct {
	*prometheus.Registry
	logger *observability.Logger
}

// Server is a metrics server
type Server struct {
	server   *http.Server
	registry *Registry
	logger   *observability.Logger
	options  *ServerOptions
}

// ServerOptions contains options for the metrics server
type ServerOptions struct {
	Host string
	Port int
	Path string
}

// DefaultServerOptions returns the default server options
func DefaultServerOptions() *ServerOptions {
	return &ServerOptions{
		Host: "0.0.0.0",
		Port: 9100,
		Path: "/metrics",
	}
}

// NewRegistry creates a new metrics registry
func NewRegistry(logger *observability.Logger) *Registry {
	registry := prometheus.NewRegistry()

	// Register standard collectors
	registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	registry.MustRegister(prometheus.NewGoCollector())

	return &Registry{
		Registry: registry,
		logger:   logger,
	}
}

// NewServer creates a new metrics server
func NewServer(registry *Registry, logger *observability.Logger, options *ServerOptions) *Server {
	if options == nil {
		options = DefaultServerOptions()
	}

	// Create handler
	handler := promhttp.HandlerFor(registry.Registry, promhttp.HandlerOpts{
		Registry: registry.Registry,
	})

	// Create mux
	mux := http.NewServeMux()
	mux.Handle(options.Path, handler)

	// Create server
	addr := fmt.Sprintf("%s:%d", options.Host, options.Port)
	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	return &Server{
		server:   server,
		registry: registry,
		logger:   logger,
		options:  options,
	}
}

// Start starts the metrics server
func (s *Server) Start() error {
	s.logger.Info("Starting metrics server", zap.String("address", s.server.Addr), zap.String("path", s.options.Path))
	return s.server.ListenAndServe()
}

// Stop stops the metrics server
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Stopping metrics server")
	return s.server.Shutdown(ctx)
}

// Counter creates a new counter
func (r *Registry) Counter(opts prometheus.CounterOpts) prometheus.Counter {
	counter := prometheus.NewCounter(opts)
	r.MustRegister(counter)
	return counter
}

// CounterVec creates a new counter vector
func (r *Registry) CounterVec(opts prometheus.CounterOpts, labelNames []string) *prometheus.CounterVec {
	counterVec := prometheus.NewCounterVec(opts, labelNames)
	r.MustRegister(counterVec)
	return counterVec
}

// Gauge creates a new gauge
func (r *Registry) Gauge(opts prometheus.GaugeOpts) prometheus.Gauge {
	gauge := prometheus.NewGauge(opts)
	r.MustRegister(gauge)
	return gauge
}

// GaugeVec creates a new gauge vector
func (r *Registry) GaugeVec(opts prometheus.GaugeOpts, labelNames []string) *prometheus.GaugeVec {
	gaugeVec := prometheus.NewGaugeVec(opts, labelNames)
	r.MustRegister(gaugeVec)
	return gaugeVec
}

// Histogram creates a new histogram
func (r *Registry) Histogram(opts prometheus.HistogramOpts) prometheus.Histogram {
	histogram := prometheus.NewHistogram(opts)
	r.MustRegister(histogram)
	return histogram
}

// HistogramVec creates a new histogram vector
func (r *Registry) HistogramVec(opts prometheus.HistogramOpts, labelNames []string) *prometheus.HistogramVec {
	histogramVec := prometheus.NewHistogramVec(opts, labelNames)
	r.MustRegister(histogramVec)
	return histogramVec
}

// Summary creates a new summary
func (r *Registry) Summary(opts prometheus.SummaryOpts) prometheus.Summary {
	summary := prometheus.NewSummary(opts)
	r.MustRegister(summary)
	return summary
}

// SummaryVec creates a new summary vector
func (r *Registry) SummaryVec(opts prometheus.SummaryOpts, labelNames []string) *prometheus.SummaryVec {
	summaryVec := prometheus.NewSummaryVec(opts, labelNames)
	r.MustRegister(summaryVec)
	return summaryVec
}

// Timer is a helper for timing operations
type Timer struct {
	histogram prometheus.Observer
	start     time.Time
}

// NewTimer creates a new timer
func NewTimer(histogram prometheus.Observer) *Timer {
	return &Timer{
		histogram: histogram,
		start:     time.Now(),
	}
}

// ObserveDuration observes the duration since the timer was created
func (t *Timer) ObserveDuration() {
	t.histogram.Observe(time.Since(t.start).Seconds())
}
