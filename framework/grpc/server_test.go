package grpc

import (
	"context"
	"testing"
	"time"

	"github.com/axiomod/axiomod/framework/config"
	"github.com/axiomod/axiomod/framework/observability"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// testLogger returns a no-op observability logger.
func testLogger() *observability.Logger {
	return &observability.Logger{Logger: zap.NewNop()}
}

// testMetrics returns a metrics instance backed by a fresh registry.
func testMetrics(t *testing.T) *observability.Metrics {
	t.Helper()
	cfg := &config.Config{
		Observability: config.ObservabilityConfig{MetricsEnabled: true},
	}
	metrics, err := observability.NewMetrics(cfg, testLogger())
	require.NoError(t, err)
	return metrics
}

// testTracer returns a tracer backed by a no-op tracer provider.
func testTracer() *observability.Tracer {
	return &observability.Tracer{
		Tracer: tracenoop.NewTracerProvider().Tracer("test"),
	}
}

func TestNewServerOptions(t *testing.T) {
	cfg := &config.Config{
		GRPC: config.GRPCConfig{
			Host: "127.0.0.1",
			Port: 9999,
		},
	}

	opts := NewServerOptions(cfg)
	require.NotNil(t, opts)
	assert.Equal(t, "127.0.0.1", opts.Host)
	assert.Equal(t, 9999, opts.Port)
	assert.Equal(t, time.Hour, opts.MaxConnectionAge)
	assert.Equal(t, 15*time.Minute, opts.MaxConnectionIdle)
	assert.Equal(t, 30*time.Second, opts.Timeout)
	assert.Empty(t, opts.TLSCertFile)
	assert.Empty(t, opts.TLSKeyFile)
	assert.Nil(t, opts.AuthFunc)
}

func TestDefaultServerOptions(t *testing.T) {
	opts := DefaultServerOptions()
	require.NotNil(t, opts)
	assert.Equal(t, "0.0.0.0", opts.Host)
	assert.Equal(t, 9090, opts.Port)
	assert.Equal(t, time.Hour, opts.MaxConnectionAge)
	assert.Equal(t, 15*time.Minute, opts.MaxConnectionIdle)
	assert.Equal(t, 30*time.Second, opts.Timeout)
	assert.Nil(t, opts.AuthFunc)
}

func TestNewServer(t *testing.T) {
	options := &ServerOptions{
		Host:              "127.0.0.1",
		Port:              0, // Let the OS pick a free port.
		MaxConnectionAge:  time.Hour,
		MaxConnectionIdle: 15 * time.Minute,
		Timeout:           time.Second,
	}

	server, err := NewServer(testLogger(), options, NewMetricsInterceptor(testMetrics(t)), NewTracingInterceptor(testTracer()))
	require.NoError(t, err)
	require.NotNil(t, server)
	assert.NotNil(t, server.GetServer())

	server.Stop()
}

func TestNewServer_StartAndStop(t *testing.T) {
	options := &ServerOptions{
		Host:              "127.0.0.1",
		Port:              0,
		MaxConnectionAge:  time.Hour,
		MaxConnectionIdle: 15 * time.Minute,
		Timeout:           time.Second,
	}

	server, err := NewServer(testLogger(), options, NewMetricsInterceptor(testMetrics(t)), NewTracingInterceptor(testTracer()))
	require.NoError(t, err)

	served := make(chan error, 1)
	go func() {
		served <- server.Start()
	}()

	// Give the server a moment to start serving, then stop it gracefully.
	time.Sleep(200 * time.Millisecond)
	server.Stop()

	select {
	case err := <-served:
		assert.NoError(t, err, "graceful stop should make Serve return nil")
	case <-time.After(10 * time.Second):
		t.Fatal("server did not stop in time")
	}
}

func TestNewServer_TLSError(t *testing.T) {
	options := &ServerOptions{
		Host:        "127.0.0.1",
		Port:        0,
		Timeout:     time.Second,
		TLSCertFile: "/nonexistent/cert.pem",
		TLSKeyFile:  "/nonexistent/key.pem",
	}

	server, err := NewServer(testLogger(), options, NewMetricsInterceptor(testMetrics(t)), NewTracingInterceptor(testTracer()))
	require.Error(t, err)
	assert.Nil(t, server)
	assert.Contains(t, err.Error(), "failed to load TLS credentials")
}

func TestRecoveryHandler(t *testing.T) {
	handler := recoveryHandler(testLogger())

	err := handler("something went wrong")
	require.Error(t, err)

	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
	assert.Equal(t, "internal server error", st.Message())
}

func TestTimeoutInterceptor_CompletesBeforeTimeout(t *testing.T) {
	interceptor := timeoutInterceptor(time.Second)
	info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}

	resp, err := interceptor(context.Background(), "request", info, func(ctx context.Context, req interface{}) (interface{}, error) {
		assert.Equal(t, "request", req)
		_, hasDeadline := ctx.Deadline()
		assert.True(t, hasDeadline, "handler context must carry the timeout deadline")
		return "response", nil
	})

	require.NoError(t, err)
	assert.Equal(t, "response", resp)
}

func TestTimeoutInterceptor_HandlerErrorPassthrough(t *testing.T) {
	interceptor := timeoutInterceptor(time.Second)
	info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}
	handlerErr := status.Error(codes.NotFound, "missing")

	resp, err := interceptor(context.Background(), "request", info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, handlerErr
	})

	assert.Nil(t, resp)
	assert.Equal(t, handlerErr, err)
}

func TestTimeoutInterceptor_TimesOut(t *testing.T) {
	interceptor := timeoutInterceptor(100 * time.Millisecond)
	info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}

	block := make(chan struct{})
	defer close(block)

	start := time.Now()
	resp, err := interceptor(context.Background(), "request", info, func(ctx context.Context, req interface{}) (interface{}, error) {
		<-block
		return "too late", nil
	})

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "request timeout")
	assert.Less(t, time.Since(start), 5*time.Second)
}
