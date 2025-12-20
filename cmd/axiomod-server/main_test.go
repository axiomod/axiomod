package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/axiomod/axiomod/framework/config"
	"github.com/axiomod/axiomod/framework/auth"
	"github.com/axiomod/axiomod/framework/cache"
	"github.com/axiomod/axiomod/framework/circuitbreaker"
	"github.com/axiomod/axiomod/framework/worker"
	"github.com/axiomod/axiomod/platform/observability"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// Helper function to create a temporary config file for testing
func createTestConfigFile(t *testing.T, content string) string {
	 t.Helper()
	 dir := t.TempDir()
	 path := filepath.Join(dir, "test_config.yaml")
	 err := os.WriteFile(path, []byte(content), 0644)
	 assert.NoError(t, err)
	 return path
}

func TestConfigLoading(t *testing.T) {
	 // Create a dummy config file for testing defaults
	 testConfigContent := `
app:
  name: test-app-from-file
  environment: test-env
  version: 1.1.0
http:
  port: 9090
`
	 testConfigPath := createTestConfigFile(t, testConfigContent)

	 // Load config from the test file
	 cfg, err := config.Load(testConfigPath)
	 assert.NoError(t, err)
	 assert.NotNil(t, cfg)

	 // Check values loaded from the test file
	 assert.Equal(t, "test-app-from-file", cfg.App.Name)
	 assert.Equal(t, "test-env", cfg.App.Environment)
	 assert.Equal(t, 9090, cfg.HTTP.Port)

	 // Test loading with empty path (should load defaults from service_default.yaml)
	 // Ensure service_default.yaml exists in a location Viper checks (e.g., ./framework/config)
	 cfgDefault, errDefault := config.Load("")
	 assert.NoError(t, errDefault)
	 assert.NotNil(t, cfgDefault)

	 // Check default values loaded from service_default.yaml
	 assert.Equal(t, "axiomod-default", cfgDefault.App.Name, "Default App.Name mismatch")
	 assert.Equal(t, "development", cfgDefault.App.Environment, "Default App.Environment mismatch")
	 assert.Equal(t, 8080, cfgDefault.HTTP.Port, "Default HTTP.Port mismatch")
}

func TestJWTAuth(t *testing.T) {
	 // Create JWT service
	 jwtService := auth.NewJWTService("test-secret", time.Hour)

	 // Generate token
	 token, err := jwtService.GenerateToken("user123", "testuser", "test@example.com", []string{"admin"})
	 assert.NoError(t, err)
	 assert.NotEmpty(t, token)

	 // Validate token
	 claims, err := jwtService.ValidateToken(token)
	 assert.NoError(t, err)
	 assert.Equal(t, "user123", claims.UserID)
	 assert.Equal(t, "testuser", claims.Username)
	 assert.Equal(t, "test@example.com", claims.Email)
	 assert.Contains(t, claims.Roles, "admin")
	 assert.True(t, claims.HasRole("admin"))
	 assert.False(t, claims.HasRole("user"))
}

func TestMemoryCache(t *testing.T) {
	 // Create memory cache
	 memCache := cache.NewMemoryCache(100)
	 ctx := context.Background()

	 // Set value
	 err := memCache.Set(ctx, "test-key", []byte("test-value"), time.Minute)
	 assert.NoError(t, err)

	 // Get value
	 value, err := memCache.Get(ctx, "test-key")
	 assert.NoError(t, err)
	 assert.Equal(t, []byte("test-value"), value)

	 // Delete value
	 err = memCache.Delete(ctx, "test-key")
	 assert.NoError(t, err)

	 // Get non-existent value
	 _, err = memCache.Get(ctx, "test-key")
	 assert.Error(t, err)
	 assert.Equal(t, cache.ErrKeyNotFound, err)
}

func TestCircuitBreaker(t *testing.T) {
	 // Create circuit breaker
	 cb := circuitbreaker.New(circuitbreaker.Options{
		 Name:          "test",
		 MaxFailures:   2,
		 ResetTimeout:  50 * time.Millisecond,
		 HalfOpenLimit: 1,
	 })

	 // Test successful execution
	 err := cb.Execute(func() error {
		 return nil
	 })
	 assert.NoError(t, err)
	 assert.Equal(t, circuitbreaker.StateClosed, cb.State())

	 // Test failed execution
	 testErr := errors.New("test error")
	 err = cb.Execute(func() error {
		 return testErr
	 })
	 assert.Equal(t, testErr, err)
	 assert.Equal(t, circuitbreaker.StateClosed, cb.State())

	 // Test circuit breaker opening
	 err = cb.Execute(func() error {
		 return testErr
	 })
	 assert.Equal(t, testErr, err)
	 assert.Equal(t, circuitbreaker.StateOpen, cb.State())

	 // Test circuit breaker rejecting requests
	 err = cb.Execute(func() error {
		 return nil
	 })
	 assert.Error(t, err)
	 assert.Contains(t, err.Error(), "circuit breaker is open")

	 // Wait for reset timeout
	 time.Sleep(60 * time.Millisecond)

	 // Test half-open state - successful request should close the circuit
	 err = cb.Execute(func() error {
		 return nil
	 })
	 assert.NoError(t, err)
	 assert.Equal(t, circuitbreaker.StateClosed, cb.State())

	 // Test half-open state - failed request should re-open the circuit
	 // First, open the circuit again
	 cb.Execute(func() error { return testErr })
	 cb.Execute(func() error { return testErr })
	 assert.Equal(t, circuitbreaker.StateOpen, cb.State())
	 time.Sleep(60 * time.Millisecond) // Wait for reset
	 // AllowRequest should transition state to HalfOpen implicitly
	 cb.AllowRequest() // Trigger potential state transition
	 assert.Equal(t, circuitbreaker.StateHalfOpen, cb.State(), "Should be HalfOpen after timeout")
	 err = cb.Execute(func() error {
		 return testErr
	 })
	 assert.Equal(t, testErr, err)
	 assert.Equal(t, circuitbreaker.StateOpen, cb.State(), "Should be Open after failure in HalfOpen") // Expect StateOpen (1) after failure in HalfOpen
}

func TestWorker(t *testing.T) {
	 // Create logger
	 logger, _ := zap.NewDevelopment()
	 obsLogger := &observability.Logger{Logger: logger}

	 // Create worker
	 w := worker.New(obsLogger)

	 // Create job
	 jobExecuted := false
	 job := &worker.Job{
		 ID:       "test-job",
		 Name:     "Test Job",
		 Interval: 50 * time.Millisecond, // Faster interval for testing
		 Timeout:  time.Second,
		 Func: func(ctx context.Context) error {
			 jobExecuted = true
			 return nil
		 },
	 }

	 // Register job
	 err := w.RegisterJob(job)
	 assert.NoError(t, err)

	 // Start job
	 err = w.StartJob(job.ID)
	 assert.NoError(t, err)

	 // Wait longer to ensure job executes
	 time.Sleep(150 * time.Millisecond)

	 // Stop job
	 err = w.StopJob(job.ID)
	 assert.NoError(t, err)

	 // Check if job was executed
	 assert.True(t, jobExecuted)
}
