package grpc

import (
	"context"
	"testing"

	"github.com/axiomod/axiomod/framework/observability"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// counterValue reads the current value of a prometheus counter.
func counterValue(t *testing.T, counter prometheus.Counter) float64 {
	t.Helper()
	var metric dto.Metric
	require.NoError(t, counter.Write(&metric))
	return metric.GetCounter().GetValue()
}

// gatherFamilyNames returns the set of metric family names in the registry.
func gatherFamilyNames(t *testing.T, metrics *observability.Metrics) map[string]bool {
	t.Helper()
	families, err := metrics.Registry.Gather()
	require.NoError(t, err)
	names := make(map[string]bool, len(families))
	for _, family := range families {
		names[family.GetName()] = true
	}
	return names
}

func TestParseFullMethod(t *testing.T) {
	tests := []struct {
		name            string
		fullMethod      string
		expectedService string
		expectedMethod  string
	}{
		{"standard method", "/pkg.Service/Method", "pkg.Service", "Method"},
		{"nested package", "/a.b.c.Service/Do", "a.b.c.Service", "Do"},
		{"empty string", "", "unknown", ""},
		{"missing leading slash", "pkg.Service/Method", "unknown", "pkg.Service/Method"},
		{"no method separator", "/justservice", "unknown", "justservice"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, method := parseFullMethod(tt.fullMethod)
			assert.Equal(t, tt.expectedService, service)
			assert.Equal(t, tt.expectedMethod, method)
		})
	}
}

func TestNewMetricsInterceptor(t *testing.T) {
	interceptor := NewMetricsInterceptor(testMetrics(t))
	require.NotNil(t, interceptor)
	assert.NotNil(t, interceptor.Unary())
}

func TestMetricsInterceptor_RecordsSuccess(t *testing.T) {
	metrics := testMetrics(t)
	interceptor := NewMetricsInterceptor(metrics).Unary()
	info := &grpc.UnaryServerInfo{FullMethod: "/pkg.Service/Method"}

	resp, err := interceptor(context.Background(), "request", info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return "response", nil
	})

	require.NoError(t, err)
	assert.Equal(t, "response", resp)

	counter := metrics.GRPCRequestsTotal.WithLabelValues("pkg.Service", "Method", codes.OK.String())
	assert.Equal(t, float64(1), counterValue(t, counter))

	names := gatherFamilyNames(t, metrics)
	assert.True(t, names["grpc_requests_total"])
	assert.True(t, names["grpc_request_duration_seconds"])
}

func TestMetricsInterceptor_RecordsErrorStatus(t *testing.T) {
	metrics := testMetrics(t)
	interceptor := NewMetricsInterceptor(metrics).Unary()
	info := &grpc.UnaryServerInfo{FullMethod: "/pkg.Service/Method"}
	handlerErr := status.Error(codes.NotFound, "missing")

	resp, err := interceptor(context.Background(), "request", info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, handlerErr
	})

	assert.Nil(t, resp)
	assert.Equal(t, handlerErr, err)

	counter := metrics.GRPCRequestsTotal.WithLabelValues("pkg.Service", "Method", codes.NotFound.String())
	assert.Equal(t, float64(1), counterValue(t, counter))
}

func TestMetricsInterceptor_SkipsHealthAndReflection(t *testing.T) {
	tests := []struct {
		name       string
		fullMethod string
	}{
		{"health check", "/grpc.health.v1.Health/Check"},
		{"reflection", "/grpc.reflection.v1alpha.ServerReflection/ServerReflectionInfo"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := testMetrics(t)
			interceptor := NewMetricsInterceptor(metrics).Unary()
			info := &grpc.UnaryServerInfo{FullMethod: tt.fullMethod}

			resp, err := interceptor(context.Background(), "request", info, func(ctx context.Context, req interface{}) (interface{}, error) {
				return "response", nil
			})

			require.NoError(t, err)
			assert.Equal(t, "response", resp)
			// No grpc metric must have been recorded for the skipped method.
			names := gatherFamilyNames(t, metrics)
			assert.False(t, names["grpc_requests_total"])
			assert.False(t, names["grpc_request_duration_seconds"])
		})
	}
}
