package main_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/axiomod/axiomod/examples/example/delivery/http/middleware"
	"github.com/axiomod/axiomod/platform/observability"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// TestIntegration is a simple integration test that starts the HTTP server
// and makes a request to the health endpoint.
func TestIntegration(t *testing.T) {
	// Create context with cancellation
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	// Use a fixed port for testing instead of relying on config
	testPort := 8099
	testHost := "127.0.0.1"

	// Create logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	obsLogger := &observability.Logger{Logger: logger}

	// Create Fiber app with specific config for testing
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	// Add middleware
	loggingMiddleware := middleware.NewLoggingMiddleware(obsLogger)
	app.Use(loggingMiddleware.Handle())

	// Add health endpoint
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(map[string]string{"status": "ok"})
	})

	// Start server in a goroutine with fixed address
	serverAddr := fmt.Sprintf("%s:%d", testHost, testPort)
	go func() {
		obsLogger.Info("Starting HTTP server", zap.String("address", serverAddr))
		if err := app.Listen(serverAddr); err != nil && err != http.ErrServerClosed {
			obsLogger.Error("Failed to start HTTP server", zap.Error(err))
		}
	}()

	// Wait for server to start by checking if the port is open
	waitForServer := func(addr string, timeout time.Duration) bool {
		deadline := time.Now().Add(timeout)
		for time.Now().Before(deadline) {
			conn, err := net.DialTimeout("tcp", addr, 100*time.Millisecond)
			if err == nil {
				conn.Close()
				return true
			}
			time.Sleep(100 * time.Millisecond)
		}
		return false
	}

	if !waitForServer(serverAddr, 3*time.Second) {
		t.Fatalf("Server failed to start within timeout")
	}

	// Make a request to the health endpoint
	resp, err := http.Get(fmt.Sprintf("http://%s/health", serverAddr))
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Check response
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected status code 200")

	obsLogger.Info("Integration test passed")

	// Shutdown server
	if err := app.Shutdown(); err != nil {
		t.Fatalf("Failed to shutdown server: %v", err)
	}
}
