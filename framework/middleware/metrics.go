package middleware

import (
	"strconv"
	"time"

	"github.com/axiomod/axiomod/platform/observability"
	"github.com/gofiber/fiber/v2"
)

// MetricsMiddleware records HTTP request metrics
type MetricsMiddleware struct {
	metrics *observability.Metrics
}

// NewMetricsMiddleware creates a new metrics middleware
func NewMetricsMiddleware(metrics *observability.Metrics) *MetricsMiddleware {
	return &MetricsMiddleware{
		metrics: metrics,
	}
}

// Handle returns a Fiber middleware handler
func (m *MetricsMiddleware) Handle() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Skip metrics for the /metrics path to avoid noise
		if c.Path() == "/metrics" {
			return c.Next()
		}

		err := c.Next()

		status := strconv.Itoa(c.Response().StatusCode())
		method := c.Method()
		path := c.Path()

		// Use the route path if available to avoid high cardinality
		if route := c.Route(); route != nil {
			path = route.Path
		}

		duration := time.Since(start).Seconds()

		if m.metrics.HTTPRequestsTotal != nil {
			m.metrics.HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()
		}
		if m.metrics.HTTPRequestDuration != nil {
			m.metrics.HTTPRequestDuration.WithLabelValues(method, path, status).Observe(duration)
		}

		return err
	}
}
