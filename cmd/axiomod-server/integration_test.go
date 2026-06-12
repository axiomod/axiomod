package main_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/axiomod/axiomod/examples/example"
	"github.com/axiomod/axiomod/framework/auth"
	"github.com/axiomod/axiomod/framework/config"
	grpc_pkg "github.com/axiomod/axiomod/framework/grpc"
	"github.com/axiomod/axiomod/framework/health"
	"github.com/axiomod/axiomod/framework/middleware"
	"github.com/axiomod/axiomod/framework/observability"
	"github.com/axiomod/axiomod/framework/worker"
	"github.com/axiomod/axiomod/platform/server"
	"github.com/axiomod/axiomod/plugins"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
)

// testJWTSecret satisfies the framework's 32-byte minimum for HMAC signing.
const testJWTSecret = "integration-test-secret-0123456789abcdef"

// testConfig returns a config that boots the full module graph with no
// external infrastructure: no plugins enabled, tracing off, fixed test ports.
func testConfig(httpPort, grpcPort int) *config.Config {
	cfg := &config.Config{}
	cfg.App.Name = "integration-test"
	cfg.App.Environment = "test"
	cfg.Observability.LogLevel = "error"
	cfg.Observability.LogFormat = "console"
	cfg.HTTP.Host = "127.0.0.1"
	cfg.HTTP.Port = httpPort
	cfg.GRPC.Host = "127.0.0.1"
	cfg.GRPC.Port = grpcPort
	cfg.Auth.JWT.SecretKey = testJWTSecret
	cfg.Auth.JWT.TokenDuration = 60
	return cfg
}

// waitForServer polls until the TCP port accepts connections.
func waitForServer(t *testing.T, addr string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", addr, 100*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("server at %s failed to start within %s", addr, timeout)
}

// TestIntegration starts a plain Fiber app with the framework logging
// middleware and checks the health endpoint responds.
func TestIntegration(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	obsLogger := &observability.Logger{Logger: logger}

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(middleware.NewLoggingMiddleware(obsLogger).Handle())
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(map[string]string{"status": "ok"})
	})

	serverAddr := "127.0.0.1:8099"
	go func() {
		if err := app.Listen(serverAddr); err != nil {
			obsLogger.Error("Failed to start HTTP server", zap.Error(err))
		}
	}()
	waitForServer(t, serverAddr, 3*time.Second)

	resp, err := http.Get(fmt.Sprintf("http://%s/health", serverAddr))
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	require.NoError(t, app.Shutdown())
}

// TestExampleModuleCRUD boots the full server assembly (the same modules as
// fx_options.go, including the example domain module) and exercises the
// complete CRUD surface over HTTP with real JWT authentication.
func TestExampleModuleCRUD(t *testing.T) {
	const httpPort, grpcPort = 18099, 19099
	cfg := testConfig(httpPort, grpcPort)

	app := fxtest.New(t,
		fx.NopLogger,
		fx.Provide(func() *config.Config { return cfg }),
		observability.Module,
		middleware.Module,
		auth.Module,
		health.Module,
		grpc_pkg.Module,
		server.Module,
		plugins.Module,
		worker.Module,
		example.Module,
		fx.Invoke(server.RegisterHTTPServer, server.RegisterGRPCServer),
	)
	app.RequireStart()
	defer app.RequireStop()

	base := fmt.Sprintf("http://127.0.0.1:%d", httpPort)
	waitForServer(t, fmt.Sprintf("127.0.0.1:%d", httpPort), 5*time.Second)

	// Mint a token with the same secret the server validates against.
	jwtService := auth.NewJWTService(testJWTSecret, time.Hour)
	token, err := jwtService.GenerateToken("user-1", "tester", "tester@example.com", []string{"admin"})
	require.NoError(t, err)

	client := &http.Client{Timeout: 5 * time.Second}
	doJSON := func(method, path string, body interface{}, useAuth bool) (*http.Response, map[string]interface{}) {
		t.Helper()
		var reader *bytes.Reader
		if body != nil {
			raw, err := json.Marshal(body)
			require.NoError(t, err)
			reader = bytes.NewReader(raw)
		} else {
			reader = bytes.NewReader(nil)
		}
		req, err := http.NewRequestWithContext(context.Background(), method, base+path, reader)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		if useAuth {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		var decoded map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&decoded)
		return resp, decoded
	}

	// Unauthenticated requests are rejected.
	resp, _ := doJSON(http.MethodGet, "/api/v1/examples/", nil, false)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	// Create
	resp, created := doJSON(http.MethodPost, "/api/v1/examples/", map[string]interface{}{
		"name":        "integration",
		"description": "full crud round-trip",
		"valueType":   "premium",
		"count":       3,
		"tags":        []string{"a"},
	}, true)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	id, _ := created["id"].(string)
	require.NotEmpty(t, id)

	// Get
	resp, got := doJSON(http.MethodGet, "/api/v1/examples/"+id, nil, true)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "integration", got["name"])

	// Update
	resp, updated := doJSON(http.MethodPut, "/api/v1/examples/"+id, map[string]interface{}{
		"name":        "integration-updated",
		"description": "updated",
		"valueType":   "basic",
		"count":       5,
	}, true)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "integration-updated", updated["name"])

	// List
	resp, listed := doJSON(http.MethodGet, "/api/v1/examples/?valueType=basic", nil, true)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.GreaterOrEqual(t, listed["total"], float64(1))

	// Delete
	resp, _ = doJSON(http.MethodDelete, "/api/v1/examples/"+id, nil, true)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Get after delete -> 404
	resp, _ = doJSON(http.MethodGet, "/api/v1/examples/"+id, nil, true)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}
