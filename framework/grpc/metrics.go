package grpc

import (
	"context"
	"time"

	"github.com/axiomod/axiomod/platform/observability"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// MetricsInterceptor records gRPC request metrics
type MetricsInterceptor struct {
	metrics *observability.Metrics
}

// NewMetricsInterceptor creates a new metrics interceptor
func NewMetricsInterceptor(metrics *observability.Metrics) *MetricsInterceptor {
	return &MetricsInterceptor{
		metrics: metrics,
	}
}

// Unary returns a gRPC unary interceptor
func (i *MetricsInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()

		// Skip metrics for health check and reflection
		if info.FullMethod == "/grpc.health.v1.Health/Check" || info.FullMethod == "/grpc.reflection.v1alpha.ServerReflection/ServerReflectionInfo" {
			return handler(ctx, req)
		}

		resp, err := handler(ctx, req)

		st, _ := status.FromError(err)
		statusCode := st.Code().String()
		service, method := parseFullMethod(info.FullMethod)

		duration := time.Since(start).Seconds()

		i.metrics.GRPCRequestsTotal.WithLabelValues(service, method, statusCode).Inc()
		i.metrics.GRPCRequestDuration.WithLabelValues(service, method, statusCode).Observe(duration)

		return resp, err
	}
}

// parseFullMethod splits full method into service and method
func parseFullMethod(fullMethod string) (string, string) {
	// FullMethod format: /package.Service/Method
	if len(fullMethod) == 0 || fullMethod[0] != '/' {
		return "unknown", fullMethod
	}

	for i := 1; i < len(fullMethod); i++ {
		if fullMethod[i] == '/' {
			return fullMethod[1:i], fullMethod[i+1:]
		}
	}

	return "unknown", fullMethod[1:]
}
