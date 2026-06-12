package grpc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestNewTracingInterceptor(t *testing.T) {
	interceptor := NewTracingInterceptor(testTracer())
	require.NotNil(t, interceptor)
	assert.NotNil(t, interceptor.Unary())
}

func TestTracingInterceptor_Success(t *testing.T) {
	interceptor := NewTracingInterceptor(testTracer()).Unary()
	info := &grpc.UnaryServerInfo{FullMethod: "/pkg.Service/Method"}

	resp, err := interceptor(context.Background(), "request", info, func(ctx context.Context, req interface{}) (interface{}, error) {
		assert.Equal(t, "request", req)
		return "response", nil
	})

	require.NoError(t, err)
	assert.Equal(t, "response", resp)
}

func TestTracingInterceptor_WithIncomingMetadata(t *testing.T) {
	interceptor := NewTracingInterceptor(testTracer()).Unary()
	info := &grpc.UnaryServerInfo{FullMethod: "/pkg.Service/Method"}

	md := metadata.Pairs("traceparent", "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")
	ctx := metadata.NewIncomingContext(context.Background(), md)

	resp, err := interceptor(ctx, "request", info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return "response", nil
	})

	require.NoError(t, err)
	assert.Equal(t, "response", resp)
}

func TestTracingInterceptor_ErrorPassthrough(t *testing.T) {
	interceptor := NewTracingInterceptor(testTracer()).Unary()
	info := &grpc.UnaryServerInfo{FullMethod: "/pkg.Service/Method"}
	handlerErr := status.Error(codes.Unavailable, "down")

	resp, err := interceptor(context.Background(), "request", info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, handlerErr
	})

	assert.Nil(t, resp)
	assert.Equal(t, handlerErr, err)
}

func TestMetadataCarrier(t *testing.T) {
	md := metadata.Pairs("key-one", "value-one", "key-two", "value-two")
	carrier := metadataCarrier(md)

	t.Run("get existing key", func(t *testing.T) {
		assert.Equal(t, "value-one", carrier.Get("key-one"))
	})

	t.Run("get missing key", func(t *testing.T) {
		assert.Empty(t, carrier.Get("missing"))
	})

	t.Run("set key", func(t *testing.T) {
		carrier.Set("key-three", "value-three")
		assert.Equal(t, "value-three", carrier.Get("key-three"))
	})

	t.Run("keys", func(t *testing.T) {
		keys := carrier.Keys()
		assert.Contains(t, keys, "key-one")
		assert.Contains(t, keys, "key-two")
	})
}
