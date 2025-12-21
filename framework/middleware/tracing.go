package middleware

import (
	"github.com/axiomod/axiomod/platform/observability"
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// TracingMiddleware records OTel spans for HTTP requests
type TracingMiddleware struct {
	tracer *observability.Tracer
}

// NewTracingMiddleware creates a new tracing middleware
func NewTracingMiddleware(tracer *observability.Tracer) *TracingMiddleware {
	return &TracingMiddleware{
		tracer: tracer,
	}
}

// Handle returns a Fiber middleware handler
func (m *TracingMiddleware) Handle() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extract context from headers
		ctx := otel.GetTextMapPropagator().Extract(c.UserContext(), &fiberHeaderCarrier{c: c})

		// Start span
		spanName := c.Path()
		if route := c.Route(); route != nil {
			spanName = route.Path
		}

		ctx, span := m.tracer.Tracer.Start(ctx, spanName, trace.WithSpanKind(trace.SpanKindServer))
		defer span.End()

		// Add attributes
		span.SetAttributes(
			attribute.String("http.method", c.Method()),
			attribute.String("http.path", c.Path()),
			attribute.String("http.ip", c.IP()),
		)

		// Store context in Fiber
		c.SetUserContext(ctx)

		// Process request
		err := c.Next()

		// Update span with response info
		span.SetAttributes(attribute.Int("http.status_code", c.Response().StatusCode()))
		if err != nil {
			span.RecordError(err)
		}

		return err
	}
}

type fiberHeaderCarrier struct {
	c *fiber.Ctx
}

func (f *fiberHeaderCarrier) Get(key string) string {
	return f.c.Get(key)
}

func (f *fiberHeaderCarrier) Set(key string, value string) {
	f.c.Set(key, value)
}

func (f *fiberHeaderCarrier) Keys() []string {
	keys := make([]string, 0)
	f.c.Request().Header.VisitAll(func(k, v []byte) {
		keys = append(keys, string(k))
	})
	return keys
}
