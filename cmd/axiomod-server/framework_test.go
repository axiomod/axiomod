package main

import (
	"context"
	"testing"
	"time"

	"axiomod/internal/framework/config"
	"axiomod/internal/framework/di"
	"axiomod/internal/framework/router"
	"axiomod/internal/platform/observability"

	"github.com/stretchr/testify/assert"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

func TestFrameworkInitialization(t *testing.T) {
	// Create a test application with a short timeout
	 testApp := fxtest.New(
		 t,
		 fx.Provide(
			 func() *config.Config {
				 // Use the updated config structure with HTTP and GRPC fields
				 return &config.Config{
					 App: config.AppConfig{
						 Name:    "test-app",
						 Version: "1.0.0",
					 },
					 HTTP: config.HTTPConfig{
						 Host: "localhost",
						 Port: 8080,
					 },
					 // Add other necessary config fields if needed for the test
				 }
			 },
			 observability.NewLogger,
			 observability.NewTracer,
			 observability.NewMetrics,
			 // Provide router config
			 func() *router.Config {
				 return router.DefaultConfig()
			 },
			 router.New,
		 ),
		 fx.Invoke(func(
			 logger *observability.Logger,
			 tracer *observability.Tracer,
			 metrics *observability.Metrics,
			 router *router.Router,
		 ) {
			 // Verify that all components are initialized
			 assert.NotNil(t, logger)
			 assert.NotNil(t, tracer)
			 assert.NotNil(t, metrics)
			 assert.NotNil(t, router)
		 }),
		 fx.StartTimeout(5*time.Second),
	 )

	 // Start the application
	 ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	 defer cancel()

	 // Verify that the application starts successfully
	 err := testApp.Start(ctx)
	 assert.NoError(t, err)

	 // Stop the application
	 err = testApp.Stop(ctx)
	 assert.NoError(t, err)
}

func TestDIModule(t *testing.T) {
	 // Create a test module
	 module := di.NewModule("test-module")

	 // Add a provider
	 type testService struct {
		 Name string
	 }

	 module.Provide(func() *testService {
		 return &testService{Name: "test"}
	 })

	 // Build the module
	 option := module.Build()
	 assert.NotNil(t, option)

	 // Create a test application with the module
	 testApp := fxtest.New(
		 t,
		 option,
		 fx.Invoke(func(service *testService) {
			 // Verify that the service is initialized
			 assert.NotNil(t, service)
			 assert.Equal(t, "test", service.Name)
		 }),
	 )

	 // Start the application
	 ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	 defer cancel()

	 // Verify that the application starts successfully
	 err := testApp.Start(ctx)
	 assert.NoError(t, err)

	 // Stop the application
	 err = testApp.Stop(ctx)
	 assert.NoError(t, err)
}
