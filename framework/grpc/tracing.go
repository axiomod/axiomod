package grpc

import (
	"context"

	"github.com/axiomod/axiomod/platform/observability"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// TracingInterceptor records OTel spans for gRPC requests
type TracingInterceptor struct {
	tracer *observability.Tracer
}

// NewTracingInterceptor creates a new tracing interceptor
func NewTracingInterceptor(tracer *observability.Tracer) *TracingInterceptor {
	return &TracingInterceptor{
		tracer: tracer,
	}
}

// Unary returns a gRPC unary interceptor
func (i *TracingInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Extract context from metadata
		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			ctx = otel.GetTextMapPropagator().Extract(ctx, metadataCarrier(md))
		}

		// Start span
		service, method := parseFullMethod(info.FullMethod)
		ctx, span := i.tracer.Tracer.Start(ctx, info.FullMethod, trace.WithSpanKind(trace.SpanKindServer))
		defer span.End()

		// Add attributes
		span.SetAttributes(
			attribute.String("rpc.system", "grpc"),
			attribute.String("rpc.service", service),
			attribute.String("rpc.method", method),
		)

		resp, err := handler(ctx, req)

		// Update span with status
		st, _ := status.FromError(err)
		span.SetAttributes(attribute.String("rpc.grpc.status_code", st.Code().String()))
		if err != nil {
			span.RecordError(err)
		}

		return resp, err
	}
}

type metadataCarrier metadata.MD

func (m metadataCarrier) Get(key string) string {
	values := metadata.MD(m).Get(key)
	if len(values) > 0 {
		return values[0]
	}
	return ""
}

func (m metadataCarrier) Set(key string, value string) {
	metadata.MD(m).Set(key, value)
}

func (m metadataCarrier) Keys() []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
